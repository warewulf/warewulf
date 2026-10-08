// Package siftest builds SIF images for tests with squashfs-tools.
package siftest

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/apptainer/sif/v2/pkg/sif"
	"github.com/stretchr/testify/require"
)

// Squash returns a squashfs image of the directory src. The test is skipped
// if mksquashfs or unsquashfs is not installed.
func Squash(t *testing.T, src string) []byte {
	t.Helper()
	for _, tool := range []string{"mksquashfs", "unsquashfs"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	dst := filepath.Join(t.TempDir(), "image.squashfs")
	out, err := exec.Command("mksquashfs", src, dst, "-noappend", "-no-progress", "-all-root").CombinedOutput()
	require.NoError(t, err, string(out))
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	return data
}

// Partition returns a SIF partition descriptor for data.
func Partition(t *testing.T, data []byte, fsType sif.FSType, partType sif.PartType) sif.DescriptorInput {
	t.Helper()
	di, err := sif.NewDescriptorInput(sif.DataPartition, bytes.NewReader(data),
		sif.OptPartitionMetadata(fsType, partType, runtime.GOARCH))
	require.NoError(t, err)
	return di
}

// Write writes a SIF image with the given descriptors to path.
func Write(t *testing.T, path string, inputs ...sif.DescriptorInput) {
	t.Helper()
	f, err := sif.CreateContainerAtPath(path, sif.OptCreateWithDescriptors(inputs...))
	require.NoError(t, err)
	require.NoError(t, f.UnloadContainer())
}

// WriteSIF writes a SIF image to path whose primary system partition is a
// squashfs image containing files, which maps paths to contents.
func WriteSIF(t *testing.T, path string, files map[string]string) {
	t.Helper()
	src := t.TempDir()
	for name, content := range files {
		p := filepath.Join(src, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o755))
	}
	Write(t, path, Partition(t, Squash(t, src), sif.FsSquash, sif.PartPrimSys))
}
