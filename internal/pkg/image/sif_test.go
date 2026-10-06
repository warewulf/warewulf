package image

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/apptainer/sif/v2/pkg/sif"
	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/image/siftest"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_IsSIF(t *testing.T) {
	dir := t.TempDir()
	sifPath := filepath.Join(dir, "image.sif")
	siftest.Write(t, sifPath, siftest.Partition(t, []byte("data"), sif.FsSquash, sif.PartPrimSys))
	tarPath := filepath.Join(dir, "image.tar")
	assert.NoError(t, os.WriteFile(tarPath, bytes.Repeat([]byte{0}, 1024), 0o644))
	shortPath := filepath.Join(dir, "short")
	assert.NoError(t, os.WriteFile(shortPath, []byte("#!/usr/bin/env run-singularity\n"), 0o644))
	fifoPath := filepath.Join(dir, "fifo")
	assert.NoError(t, syscall.Mkfifo(fifoPath, 0o644))

	tests := map[string]struct {
		path  string
		isSIF bool
		err   bool
	}{
		"sif":     {path: sifPath, isSIF: true},
		"tar":     {path: tarPath},
		"short":   {path: shortPath},
		"fifo":    {path: fifoPath},
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
	squashDir := func(t *testing.T, build func(dir string)) []byte {
		dir := t.TempDir()
		build(dir)
		return siftest.Squash(t, dir)
	}
	squash := squashDir(t, func(dir string) {
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, "etc"), 0o755))
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, "usr/bin"), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "bin/sh"), []byte("shell"), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "etc/os-release"), []byte("NAME=test\n"), 0o644))
		assert.NoError(t, os.Symlink("../../bin/sh", filepath.Join(dir, "usr/bin/sh")))
	})
	noShell := squashDir(t, func(dir string) {
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, "etc"), 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "etc/os-release"), nil, 0o644))
	})
	danglingShell := squashDir(t, func(dir string) {
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
		assert.NoError(t, os.Symlink("/nonexistent/busybox", filepath.Join(dir, "bin/sh")))
	})
	// /bin/sh must not be looked up through the host's /bin
	hostShell := squashDir(t, func(dir string) {
		assert.NoError(t, os.Symlink("/bin", filepath.Join(dir, "bin")))
	})

	primary := func(data []byte) []sif.DescriptorInput {
		return []sif.DescriptorInput{siftest.Partition(t, data, sif.FsSquash, sif.PartPrimSys)}
	}
	ociIndex, err := sif.NewDescriptorInput(sif.DataOCIRootIndex, bytes.NewReader([]byte(`{"schemaVersion":2}`)))
	assert.NoError(t, err)

	tests := map[string]struct {
		inputs []sif.DescriptorInput
		name   string
		err    string
	}{
		"squashfs":              {inputs: primary(squash)},
		"invalid name":          {inputs: primary(squash), name: "bad/name", err: "illegal characters"},
		"parent directory name": {inputs: primary(squash), name: "..", err: "illegal characters"},
		"overlay partition": {inputs: []sif.DescriptorInput{
			siftest.Partition(t, squash, sif.FsSquash, sif.PartPrimSys),
			siftest.Partition(t, make([]byte, 4096), sif.FsExt3, sif.PartOverlay),
		}},
		"no primary partition": {
			inputs: []sif.DescriptorInput{siftest.Partition(t, squash, sif.FsSquash, sif.PartData)},
			err:    "could not find the primary system partition",
		},
		"ext3 partition": {
			inputs: []sif.DescriptorInput{siftest.Partition(t, squash, sif.FsExt3, sif.PartPrimSys)},
			err:    "unsupported primary partition filesystem",
		},
		"oci-sif":        {inputs: []sif.DescriptorInput{ociIndex}, err: "OCI-SIF image, which is not supported"},
		"not squashfs":   {inputs: primary(bytes.Repeat([]byte{1}, 128)), err: "is not a squashfs filesystem"},
		"no shell":       {inputs: primary(noShell), err: "has no /bin/sh"},
		"dangling shell": {inputs: primary(danglingShell), err: "has no /bin/sh"},
		"host shell":     {inputs: primary(hostShell), err: "has no /bin/sh"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			sifPath := env.GetPath("image.sif")
			siftest.Write(t, sifPath, tt.inputs...)

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

func Test_ImportSIF_oldUnsquashfs(t *testing.T) {
	src := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(src, "bin"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(src, "bin/sh"), []byte("shell"), 0o755))
	squash := siftest.Squash(t, src)

	// A fake unsquashfs 4.3, which has no -o, runs the real one.
	unsquashfs, err := exec.LookPath("unsquashfs")
	assert.NoError(t, err)
	bin := t.TempDir()
	script := "#!/bin/sh\n" +
		"[ \"$1\" = -version ] && { echo 'unsquashfs version 4.3 (2014/05/12)'; exit 0; }\n" +
		"for a; do [ \"$a\" = -o ] && exit 99; done\n" +
		"exec " + unsquashfs + " \"$@\"\n"
	assert.NoError(t, os.WriteFile(filepath.Join(bin, "unsquashfs"), []byte(script), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	env := testenv.New(t)
	defer env.RemoveAll()
	sifPath := env.GetPath("image.sif")
	siftest.Write(t, sifPath, siftest.Partition(t, squash, sif.FsSquash, sif.PartPrimSys))

	assert.NoError(t, ImportSIF(sifPath, "testImage"))
	assert.Equal(t, "shell", env.ReadFile("/var/lib/warewulf/chroots/testImage/rootfs/bin/sh"))
	// The temporary copy of the partition is removed.
	entries, err := os.ReadDir(env.GetPath("/var/lib/warewulf/chroots/testImage"))
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
}

func Test_unsquashfsHasOffset(t *testing.T) {
	tests := map[string]bool{
		"unsquashfs version 4.3 (2014/05/12)":      false,
		"unsquashfs version 4.4 (2019/08/29)":      true,
		"unsquashfs version 4.4-git.1 (2021/1/17)": true,
		"unsquashfs version 4.6.1 (2023/03/25)":    true,
		"garbage":                                  false,
	}
	for out, want := range tests {
		t.Run(out, func(t *testing.T) {
			bin := filepath.Join(t.TempDir(), "unsquashfs")
			assert.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\necho '"+out+"'\n"), 0o755))
			assert.Equal(t, want, unsquashfsHasOffset(bin))
		})
	}
}
