package oci

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	dockerarchive "github.com/containers/image/v5/docker/archive"
	"github.com/containers/image/v5/docker/daemon"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/oci/layout"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
	imgSpecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/umoci"
	"github.com/opencontainers/umoci/oci/layer"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type pullerOpt func(*puller) error

func OptSetBlobCachePath(path string) pullerOpt {
	return func(p *puller) error {
		p.blobCachePath = path
		return nil
	}
}

func OptSetTmpDirPath(path string) pullerOpt {
	return func(p *puller) error {
		p.tmpDirPath = path
		return nil
	}
}

func OptSetSystemContext(s *types.SystemContext) pullerOpt {
	return func(p *puller) error {
		p.sysCtx = s
		return nil
	}
}

func OptSetPolicyContext(pCtx *signature.PolicyContext) pullerOpt {
	return func(p *puller) error {
		p.policyCtx = pCtx
		return nil
	}
}

type puller struct {
	id            string
	blobCachePath string
	tmpDirPath    string
	sysCtx        *types.SystemContext
	policyCtx     *signature.PolicyContext
}

func NewPuller(opts ...pullerOpt) (*puller, error) {
	p := &puller{
		// default to a sensible value, but caller should set this with opts
		blobCachePath: filepath.Join(defaultCachePath, blobPrefix),
	}

	for _, o := range opts {
		if err := o(p); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// getReference parsed the uri scheme to determine
func getReference(uri string) (types.ImageReference, error) {
	if util.IsFile(uri) {
		uri = "file://" + uri
	}
	s := strings.SplitN(uri, ":", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf("invalid uri: %q", uri)
	}

	switch s[0] {
	case "docker":
		return docker.ParseReference(s[1])
	case "docker-daemon":
		return daemon.ParseReference(strings.TrimPrefix(s[1], "//"))
	case "file":
		return dockerarchive.ParseReference(strings.TrimPrefix(s[1], "/"))
	default:
		return nil, fmt.Errorf("unknown uri scheme: %q", uri)
	}
}

// GenerateID stores and returns a unique identifier derived from the sha256sum of the image manifest
func (p *puller) GenerateID(ctx context.Context, uri string) (string, error) {
	ref, err := getReference(uri)
	if err != nil {
		return "", fmt.Errorf("unable to parse uri: %v", err)
	}

	src, err := ref.NewImageSource(ctx, p.sysCtx)
	if err != nil {
		return "", err
	}

	manifestBytes, _, err := src.GetManifest(ctx, nil)
	if err != nil {
		return "", err
	}

	p.id = fmt.Sprintf("sha256:%x", sha256.Sum256(manifestBytes))
	return p.id, nil
}

func (p *puller) Pull(ctx context.Context, uri, dst string) (err error) {
	srcRef, err := getReference(uri)
	if err != nil {
		return fmt.Errorf("unable to parse uri: %v", err)
	}
	srcImage, err := srcRef.NewImage(ctx, nil)
	if err != nil {
		wwlog.ErrOut("unable to create the image source, no manifest will be created: %s", err)
	} else {

		imgInspect, err := srcImage.Inspect(ctx)
		if err != nil {
			wwlog.ErrOut("Unable to get source manifest: %s", err)
		}
		// store the manifest of the source
		err = os.MkdirAll(path.Join(dst, "src"), 0755)
		if err != nil {
			wwlog.ErrOut("problems creating manifest src dir: %s", err)
		}
		outputData := InspectOutput{
			Name: "", // Set below if DockerReference() is known
			Tag:  imgInspect.Tag,
			// Digest is set below.
			RepoTags:      []string{}, // Possibly overridden for docker.Transport.
			Created:       imgInspect.Created,
			DockerVersion: imgInspect.DockerVersion,
			Labels:        imgInspect.Labels,
			Architecture:  imgInspect.Architecture,
			Os:            imgInspect.Os,
			Layers:        imgInspect.Layers,
			LayersData:    imgInspect.LayersData,
			Env:           imgInspect.Env,
		}
		srcManifest, _, err := srcImage.Manifest(ctx)
		if err != nil {
			wwlog.ErrOut("couldn't get manifest of source: %s", err)
		} else {
			outputData.Digest, _ = manifest.Digest(srcManifest)
		}
		if dockerRef := srcImage.Reference().DockerReference(); dockerRef != nil {
			outputData.Name = dockerRef.Name()
		}
		b, _ := json.MarshalIndent(outputData, "", "    ")
		err = os.WriteFile(path.Join(dst, "src/inspect.json"), b, 0644)
		if err != nil {
			wwlog.ErrOut("problems when writing manifest of source: %s", err)
		}
	}
	srcImage.Close()

	if err != nil {
		wwlog.ErrOut("failed to write inspect data: %s", err)
	}
	cacheRef, err := layout.ParseReference(p.blobCachePath + ":" + p.id)
	if err != nil {
		return fmt.Errorf("unable to generate local oci reference: %v", err)
	}

	// copy to cache location
	_, err = copy.Image(ctx, p.policyCtx, cacheRef, srcRef, &copy.Options{
		ReportWriter:     os.Stdout,
		SourceCtx:        p.sysCtx,
		RemoveSignatures: false,
	})
	if err != nil {
		return err
	}
	return p.pullFromCache(ctx, cacheRef, dst)
}

/*
private helper function to pull out the container from the cache
*/
func (p *puller) pullFromCache(ctx context.Context, cacheRef types.ImageReference, dst string) (err error) {
	// defaults to $TMPDIR or /tmp
	tmpDir, err := os.MkdirTemp(p.tmpDirPath, "oci-bundle-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// create an oci bundle our tmpdir to avoid issues with umoci.UnpackRootfs()
	tmpRef, err := layout.ParseReference(tmpDir + ":" + "tmp")
	if err != nil {
		return fmt.Errorf("unable to generate local oci reference: %v", err)
	}

	// copy to temporary location
	_, err = copy.Image(ctx, p.policyCtx, tmpRef, cacheRef, &copy.Options{})
	if err != nil {
		return err
	}

	tmp, err := tmpRef.NewImageSource(ctx, nil)
	if err != nil {
		return err
	}
	defer tmp.Close()

	manifestBytes, _, err := tmp.GetManifest(ctx, nil)
	if err != nil {
		return err
	}
	var manifest imgSpecs.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return fmt.Errorf("unable to unmarshall manifest json: %v", err)
	}

	eng, err := umoci.OpenLayout(tmpDir)
	if err != nil {
		return fmt.Errorf("unable to open oci layout: %v", err)
	}

	var mo layer.MapOptions
	err = layer.UnpackRootfs(ctx, eng, path.Join(dst, "rootfs"), manifest, &mo, nil, imgSpecs.Descriptor{})
	if err != nil {
		return fmt.Errorf("unable to unpack rootfs: %v", err)
	}

	return nil
}

func (p *puller) PullFromCache(ctx context.Context, inspectData InspectOutput, dst string) (err error) {
	cacheRef, err := layout.ParseReference(p.blobCachePath + ":" + inspectData.Digest.String())
	if err != nil {
		return fmt.Errorf("unable to generate local oci reference: %v", err)
	}
	err = os.MkdirAll(path.Join(dst, "src"), 0755)
	if err != nil {
		wwlog.ErrOut("problems creating manifest src dir: %s", err)
	}
	b, _ := json.MarshalIndent(inspectData, "", "    ")
	err = os.WriteFile(path.Join(dst, "src/inspect.json"), b, 0644)
	if err != nil {
		wwlog.ErrOut("failed to write inspect data: %s", err)
	}

	return p.pullFromCache(ctx, cacheRef, dst)

}
