package imprt

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

func createDummyDockerArchive(t *testing.T, path string) {
	// 1. Create layer content
	var layerBuf bytes.Buffer
	tw := tar.NewWriter(&layerBuf)

	// Add "." directory
	if err := tw.WriteHeader(&tar.Header{
		Name:     ".",
		Typeflag: tar.TypeDir,
		Mode:     0755,
		Uid:      os.Getuid(),
		Gid:      os.Getgid(),
	}); err != nil {
		t.Fatal(err)
	}

	// Add "bin" directory
	if err := tw.WriteHeader(&tar.Header{
		Name:     "bin",
		Typeflag: tar.TypeDir,
		Mode:     0755,
		Uid:      os.Getuid(),
		Gid:      os.Getgid(),
	}); err != nil {
		t.Fatal(err)
	}

	// Add /bin/sh which is often checked/needed
	hdr := &tar.Header{
		Name: "bin/sh",
		Mode: 0755,
		Size: int64(len("shell")),
		Uid:  os.Getuid(),
		Gid:  os.Getgid(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("shell")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	layerBytes := layerBuf.Bytes()

	// Calculate DiffID (SHA256 of uncompressed layer)
	layerSHA := sha256.Sum256(layerBytes)
	diffID := fmt.Sprintf("sha256:%x", layerSHA)

	// 2. config.json
	configStruct := struct {
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
		RootFS       struct {
			Type    string   `json:"type"`
			DiffIDs []string `json:"diff_ids"`
		} `json:"rootfs"`
	}{
		Architecture: "amd64",
		OS:           "linux",
	}
	configStruct.RootFS.Type = "layers"
	configStruct.RootFS.DiffIDs = []string{diffID}

	configJSON, err := json.Marshal(configStruct)
	if err != nil {
		t.Fatal(err)
	}

	// 3. manifest.json
	manifestStruct := []struct {
		Config   string   `json:"Config"`
		RepoTags []string `json:"RepoTags"`
		Layers   []string `json:"Layers"`
	}{
		{
			Config:   "config.json",
			RepoTags: []string{"test:latest"},
			Layers:   []string{"layer.tar"},
		},
	}
	manifestJSON, err := json.Marshal(manifestStruct)
	if err != nil {
		t.Fatal(err)
	}

	// 4. Create the tarball
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	archiveTw := tar.NewWriter(f)

	// Write layer.tar
	if err := archiveTw.WriteHeader(&tar.Header{Name: "layer.tar", Size: int64(len(layerBytes))}); err != nil {
		t.Fatal(err)
	}
	if _, err := archiveTw.Write(layerBytes); err != nil {
		t.Fatal(err)
	}

	// Write config.json
	if err := archiveTw.WriteHeader(&tar.Header{Name: "config.json", Size: int64(len(configJSON))}); err != nil {
		t.Fatal(err)
	}
	if _, err := archiveTw.Write(configJSON); err != nil {
		t.Fatal(err)
	}

	// Write manifest.json
	if err := archiveTw.WriteHeader(&tar.Header{Name: "manifest.json", Size: int64(len(manifestJSON))}); err != nil {
		t.Fatal(err)
	}
	if _, err := archiveTw.Write(manifestJSON); err != nil {
		t.Fatal(err)
	}

	if err := archiveTw.Close(); err != nil {
		t.Fatal(err)
	}
}

func resetFlags() {
	SetUpdate = false
	SetForce = false
	SetBuild = false
	SyncUser = false
}

func Test_CobraRunE_Import(t *testing.T) {
	t.Run("Basic Import", func(t *testing.T) {
		resetFlags()
		env := testenv.New(t)
		defer env.RemoveAll()

		archivePath := env.GetPath("test-image.tar")
		createDummyDockerArchive(t, archivePath)

		args := []string{"file://" + archivePath, "test-image"}
		err := CobraRunE(&cobra.Command{}, args)

		if err != nil {
			if strings.Contains(err.Error(), "operation not permitted") || strings.Contains(err.Error(), "chown") {
				t.Logf("Caught expected error in unprivileged environment: %v", err)
				return
			}
			t.Fatalf("Unexpected error: %v", err)
		}

		assert.True(t, util.IsDir(env.GetPath("var/lib/warewulf/chroots/test-image/rootfs")), "rootfs should exist")
	})

	t.Run("Import With Name Arg", func(t *testing.T) {
		resetFlags()
		env := testenv.New(t)
		defer env.RemoveAll()

		archivePath := env.GetPath("test-image.tar")
		createDummyDockerArchive(t, archivePath)

		args := []string{"file://" + archivePath, "custom-name"}
		err := CobraRunE(&cobra.Command{}, args)

		if err != nil {
			if strings.Contains(err.Error(), "operation not permitted") || strings.Contains(err.Error(), "chown") {
				t.Logf("Caught expected error in unprivileged environment: %v", err)
				return
			}
			t.Fatalf("Unexpected error: %v", err)
		}

		assert.True(t, util.IsDir(env.GetPath("var/lib/warewulf/chroots/custom-name/rootfs")), "rootfs with custom name should exist")
	})

	t.Run("Import Existing No Update", func(t *testing.T) {
		resetFlags()
		env := testenv.New(t)
		defer env.RemoveAll()

		archivePath := env.GetPath("test-image.tar")
		createDummyDockerArchive(t, archivePath)

		// Pre-create the chroot directory to simulate existing image
		env.MkdirAll("var/lib/warewulf/chroots/existing-image/rootfs")

		args := []string{"file://" + archivePath, "existing-image"}
		err := CobraRunE(&cobra.Command{}, args)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exists")
		assert.Contains(t, err.Error(), "specify --force, --update")
	})
	t.Run("Import Existing force", func(t *testing.T) {
		resetFlags()
		env := testenv.New(t)
		defer env.RemoveAll()

		archivePath := env.GetPath("test-image.tar")
		createDummyDockerArchive(t, archivePath)

		// Pre-create the chroot directory to simulate existing image
		env.MkdirAll("var/lib/warewulf/chroots/existing-image/rootfs")

		args := []string{"file://" + archivePath, "--force", "existing-image"}
		err := CobraRunE(&cobra.Command{}, args)
		if err != nil {
			if strings.Contains(err.Error(), "operation not permitted") || strings.Contains(err.Error(), "chown") {
				t.Logf("Caught expected error in unprivileged environment: %v", err)
				return
			}
			t.Fatalf("Unexpected error: %v", err)
		}

		assert.True(t, util.IsDir(env.GetPath("var/lib/warewulf/chroots/existing-image/rootfs")), "rootfs with custom name should exist")
	})

	t.Run("Import Existing With Update", func(t *testing.T) {
		resetFlags()
		SetUpdate = true

		env := testenv.New(t)
		defer env.RemoveAll()

		archivePath := env.GetPath("test-image.tar")
		createDummyDockerArchive(t, archivePath)

		// Pre-create the chroot directory and files
		rootfsPath := "var/lib/warewulf/chroots/existing-image/rootfs"
		binShPath := rootfsPath + "/bin/sh"
		otherFilePath := rootfsPath + "/file-kept"

		// Create file that should be overwritten
		env.WriteFile(binShPath, "old_shell")
		// Create file that should persist
		env.WriteFile(otherFilePath, "persist")

		args := []string{"file://" + archivePath, "existing-image"}
		err := CobraRunE(&cobra.Command{}, args)

		if err != nil {
			if strings.Contains(err.Error(), "operation not permitted") || strings.Contains(err.Error(), "chown") {
				t.Logf("Caught expected error in unprivileged environment: %v", err)
				return
			}
			t.Fatalf("Unexpected error: %v", err)
		}

		// Checks
		// 1. Image exists (rootfs dir)
		assert.True(t, util.IsDir(env.GetPath(rootfsPath)), "rootfs directory should exist")

		// 2. bin/sh overwritten
		// In unprivileged test, if we return early above, this won't run.
		// If we are privileged (or if the error doesn't happen), this verifies the overwrite.
		if util.IsFile(env.GetPath(binShPath)) {
			content := env.ReadFile(binShPath)
			assert.Equal(t, "shell", content, "bin/sh should be overwritten")
		} else {
			t.Error("bin/sh should exist")
		}

		// 3. other-file persists
		assert.True(t, util.IsFile(env.GetPath(otherFilePath)), "file-kept should still exist")
		content := env.ReadFile(otherFilePath)
		assert.Equal(t, "persist", content, "file-kept content should be preserved")
	})
}
