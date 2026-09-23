package image

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/apptainer/sif/v2/pkg/sif"
	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/squashfs/squashfstest"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func writeTestSIF(t *testing.T, path string, inputs ...sif.DescriptorInput) {
	f, err := sif.CreateContainerAtPath(path, sif.OptCreateWithDescriptors(inputs...))
	assert.NoError(t, err)
	assert.NoError(t, f.UnloadContainer())
}

func partitionInput(t *testing.T, data []byte, fsType sif.FSType, partType sif.PartType) sif.DescriptorInput {
	di, err := sif.NewDescriptorInput(sif.DataPartition, bytes.NewReader(data),
		sif.OptPartitionMetadata(fsType, partType, runtime.GOARCH))
	assert.NoError(t, err)
	return di
}

func Test_IsSIF(t *testing.T) {
	dir := t.TempDir()
	sifPath := filepath.Join(dir, "image.sif")
	assert.NoError(t, squashfstest.WriteSIF(sifPath, []squashfstest.Entry{{Path: "bin/sh", Mode: 0o755}}))
	tarPath := filepath.Join(dir, "image.tar")
	assert.NoError(t, os.WriteFile(tarPath, bytes.Repeat([]byte{0}, 1024), 0o644))
	shortPath := filepath.Join(dir, "short")
	assert.NoError(t, os.WriteFile(shortPath, []byte("#!/usr/bin/env run-singularity\n"), 0o644))

	tests := map[string]struct {
		path  string
		isSIF bool
		err   bool
	}{
		"sif":     {path: sifPath, isSIF: true},
		"tar":     {path: tarPath},
		"short":   {path: shortPath},
		"missing": {path: filepath.Join(dir, "missing"), err: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isSIF, err := IsSIF(tt.path)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.isSIF, isSIF)
		})
	}
}

func Test_ImportSIF(t *testing.T) {
	rootfs := []squashfstest.Entry{
		{Path: "bin/sh", Mode: 0o755, Content: []byte("shell")},
		{Path: "etc/os-release", Mode: 0o644, Content: []byte("NAME=test\n")},
		{Path: "usr/bin/sh", Mode: fs.ModeSymlink | 0o777, Target: "../../bin/sh"},
	}
	squash, err := squashfstest.Build(rootfs, squashfstest.Options{})
	assert.NoError(t, err)
	noShell, err := squashfstest.Build([]squashfstest.Entry{{Path: "etc/os-release", Mode: 0o644}}, squashfstest.Options{})
	assert.NoError(t, err)

	tests := map[string]struct {
		inputs func(t *testing.T) []sif.DescriptorInput
		name   string
		err    string
	}{
		"squashfs": {
			inputs: func(t *testing.T) []sif.DescriptorInput {
				return []sif.DescriptorInput{partitionInput(t, squash, sif.FsSquash, sif.PartPrimSys)}
			},
		},
		"invalid name": {
			inputs: func(t *testing.T) []sif.DescriptorInput {
				return []sif.DescriptorInput{partitionInput(t, squash, sif.FsSquash, sif.PartPrimSys)}
			},
			name: "bad/name",
			err:  "illegal characters",
		},
		"no primary partition": {
			inputs: func(t *testing.T) []sif.DescriptorInput {
				return []sif.DescriptorInput{partitionInput(t, squash, sif.FsSquash, sif.PartData)}
			},
			err: "could not find the primary system partition",
		},
		"ext3 partition": {
			inputs: func(t *testing.T) []sif.DescriptorInput {
				return []sif.DescriptorInput{partitionInput(t, squash, sif.FsExt3, sif.PartPrimSys)}
			},
			err: "unsupported primary partition filesystem",
		},
		"oci-sif": {
			inputs: func(t *testing.T) []sif.DescriptorInput {
				di, err := sif.NewDescriptorInput(sif.DataOCIRootIndex, bytes.NewReader([]byte(`{"schemaVersion":2}`)))
				assert.NoError(t, err)
				return []sif.DescriptorInput{di}
			},
			err: "OCI-SIF image, which is not supported",
		},
		"not squashfs": {
			inputs: func(t *testing.T) []sif.DescriptorInput {
				return []sif.DescriptorInput{partitionInput(t, bytes.Repeat([]byte{1}, 128), sif.FsSquash, sif.PartPrimSys)}
			},
			err: "not a squashfs filesystem",
		},
		"no shell": {
			inputs: func(t *testing.T) []sif.DescriptorInput {
				return []sif.DescriptorInput{partitionInput(t, noShell, sif.FsSquash, sif.PartPrimSys)}
			},
			err: "has no /bin/sh",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			sifPath := env.GetPath("image.sif")
			writeTestSIF(t, sifPath, tt.inputs(t)...)

			imageName := tt.name
			if imageName == "" {
				imageName = "testImage"
			}
			err := ImportSIF(sifPath, imageName)
			if tt.err != "" {
				assert.ErrorContains(t, err, tt.err)
				return
			}
			assert.NoError(t, err)
			rootfsPath := "/var/lib/warewulf/chroots/testImage/rootfs"
			assert.Equal(t, "shell", env.ReadFile(filepath.Join(rootfsPath, "bin/sh")))
			assert.Equal(t, "NAME=test\n", env.ReadFile(filepath.Join(rootfsPath, "etc/os-release")))
			target, err := os.Readlink(env.GetPath(filepath.Join(rootfsPath, "usr/bin/sh")))
			assert.NoError(t, err)
			assert.Equal(t, "../../bin/sh", target)
		})
	}
}
