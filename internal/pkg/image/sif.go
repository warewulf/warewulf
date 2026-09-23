package image

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/apptainer/sif/v2/pkg/sif"

	"github.com/warewulf/warewulf/internal/pkg/squashfs"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// The SIF global header begins with a 32-byte launch script followed by
// this magic value.
var sifMagic = []byte("SIF_MAGIC\x00")

const sifMagicOffset = 32

// IsSIF reports whether the file at path is a SIF image.
func IsSIF(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	buf := make([]byte, len(sifMagic))
	if _, err := f.ReadAt(buf, sifMagicOffset); err != nil {
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, err
	}
	return bytes.Equal(buf, sifMagic), nil
}

// ImportSIF imports the primary system partition of the SIF image at uri
// as the image name. Only squashfs partitions are supported.
func ImportSIF(uri string, name string) error {
	if !ValidName(name) {
		return errors.New("Image name contains illegal characters: " + name)
	}

	f, err := os.Open(uri)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := sif.LoadContainer(f, sif.OptLoadWithCloseOnUnload(false))
	if err != nil {
		return fmt.Errorf("could not load SIF image %s: %w", uri, err)
	}
	defer func() {
		if err := img.UnloadContainer(); err != nil {
			wwlog.Warn("failed to unload SIF image %s: %s", uri, err)
		}
	}()

	d, err := img.GetDescriptor(sif.WithPartitionType(sif.PartPrimSys))
	if err != nil {
		if _, ociErr := img.GetDescriptor(sif.WithDataType(sif.DataOCIRootIndex)); ociErr == nil {
			return fmt.Errorf("%s is an OCI-SIF image, which is not supported", uri)
		}
		return fmt.Errorf("could not find the primary system partition in %s: %w", uri, err)
	}
	fsType, _, arch, err := d.PartitionMetadata()
	if err != nil {
		return fmt.Errorf("could not read partition metadata in %s: %w", uri, err)
	}
	if fsType != sif.FsSquash {
		return fmt.Errorf("%s has an unsupported primary partition filesystem: %v", uri, fsType)
	}
	wwlog.Debug("Importing SIF primary partition: arch=%s offset=%d size=%d", arch, d.Offset(), d.Size())

	sq, err := squashfs.NewReader(io.NewSectionReader(f, d.Offset(), d.Size()))
	if err != nil {
		return fmt.Errorf("could not read squashfs partition in %s: %w", uri, err)
	}
	defer sq.Close()

	fullPath := RootFsDir(name)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return err
	}
	if err := sq.Extract(fullPath); err != nil {
		return err
	}

	if !util.IsFile(path.Join(fullPath, "/bin/sh")) {
		return errors.New("SIF image has no /bin/sh: " + uri)
	}
	return nil
}
