package squashfs

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/squashfs/squashfstest"
)

var testTime = time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)

func testEntries() []squashfstest.Entry {
	incompressible := make([]byte, 4096*3+100)
	for i := range incompressible {
		incompressible[i] = byte(i*7919 + i/13)
	}
	entries := []squashfstest.Entry{
		{Path: "bin/sh", Mode: 0o755, Content: []byte("shell"), ModTime: testTime},
		{Path: "bin/bash", Mode: fs.ModeSymlink | 0o777, Target: "sh"},
		{Path: "usr/bin/sudo", Mode: fs.ModeSetuid | 0o755, Content: []byte("sudo")},
		{Path: "usr/bin/sudoedit", Link: "usr/bin/sudo"},
		{Path: "tmp", Mode: fs.ModeDir | fs.ModeSticky | 0o777},
		{Path: "etc/big", Mode: 0o644, Content: bytes.Repeat([]byte("warewulf "), 2000)},
		{Path: "etc/random", Mode: 0o600, Content: incompressible},
		{Path: "etc/sparse", Mode: 0o644, Content: append(make([]byte, 8192), 'x')},
		{Path: "etc/zeros", Mode: 0o644, Content: make([]byte, 8192)},
		{Path: "etc/empty", Mode: 0o644},
		{Path: "run/fifo", Mode: fs.ModeNamedPipe | 0o644},
		{Path: "dev/null", Mode: fs.ModeDevice | fs.ModeCharDevice | 0o666, Major: 1, Minor: 3},
	}
	// enough entries to span several directory headers and metadata blocks
	for i := 0; i < 600; i++ {
		entries = append(entries, squashfstest.Entry{
			Path: fmt.Sprintf("many/file-with-a-long-name-%04d", i), Mode: 0o644, Content: []byte(fmt.Sprint(i)),
		})
	}
	return entries
}

func TestExtract(t *testing.T) {
	root := os.Geteuid() == 0
	for _, compression := range []squashfstest.Compression{squashfstest.Gzip, squashfstest.XZ, squashfstest.Zstd} {
		for _, noFragments := range []bool{false, true} {
			t.Run(fmt.Sprintf("compression %d noFragments %v", compression, noFragments), func(t *testing.T) {
				entries := testEntries()
				img, err := squashfstest.Build(entries, squashfstest.Options{Compression: compression, NoFragments: noFragments})
				if !assert.NoError(t, err) {
					return
				}
				sq, err := NewReader(bytes.NewReader(img))
				if !assert.NoError(t, err) {
					return
				}
				defer sq.Close()
				dst := t.TempDir()
				if !assert.NoError(t, sq.Extract(dst)) {
					return
				}

				for _, e := range entries {
					p := filepath.Join(dst, e.Path)
					if e.Mode&fs.ModeDevice != 0 && !root {
						assert.NoFileExists(t, p)
						continue
					}
					fi, err := os.Lstat(p)
					if !assert.NoError(t, err, e.Path) {
						continue
					}
					switch {
					case e.Link != "":
						assert.True(t, os.SameFile(fi, mustLstat(t, filepath.Join(dst, e.Link))), e.Path)
					case e.Mode&fs.ModeSymlink != 0:
						target, err := os.Readlink(p)
						assert.NoError(t, err)
						assert.Equal(t, e.Target, target)
					case e.Mode.IsRegular():
						content, err := os.ReadFile(p)
						assert.NoError(t, err)
						assert.Equal(t, len(e.Content), len(content), e.Path)
						assert.True(t, bytes.Equal(e.Content, content), e.Path)
						assert.Equal(t, e.Mode, fi.Mode(), e.Path)
					default:
						assert.Equal(t, e.Mode, fi.Mode(), e.Path)
					}
				}
				assert.Equal(t, testTime.Unix(), mustLstat(t, filepath.Join(dst, "bin/sh")).ModTime().Unix())
			})
		}
	}
}

func TestExtractReplacesExisting(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{
		{Path: "bin/sh", Mode: 0o755, Content: []byte("shell")},
		{Path: "etc", Mode: fs.ModeDir | 0o755},
	}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	dst := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(dst, "bin/sh/dir"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dst, "etc"), []byte("file"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dst, "kept"), []byte("kept"), 0o644))

	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	defer sq.Close()
	assert.NoError(t, sq.Extract(dst))

	content, err := os.ReadFile(filepath.Join(dst, "bin/sh"))
	assert.NoError(t, err)
	assert.Equal(t, "shell", string(content))
	assert.DirExists(t, filepath.Join(dst, "etc"))
	assert.FileExists(t, filepath.Join(dst, "kept"))
}

func TestXattrs(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{
		{Path: "file", Mode: 0o644, Xattrs: map[string][]byte{
			"user.one":            []byte("1"),
			"security.capability": {1, 2, 3},
			"trusted.three":       []byte("three"),
		}},
		{Path: "plain", Mode: 0o644},
	}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	defer sq.Close()
	root, err := sq.readInode(sq.sb.RootInode)
	assert.NoError(t, err)
	entries, err := sq.readDir(root)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	file, err := sq.readInode(entries[0].inodeRef)
	assert.NoError(t, err)
	xattrs, err := sq.xattrs(file.xattr)
	assert.NoError(t, err)
	assert.Equal(t, map[string][]byte{
		"user.one":            []byte("1"),
		"security.capability": {1, 2, 3},
		"trusted.three":       []byte("three"),
	}, xattrs)

	plain, err := sq.readInode(entries[1].inodeRef)
	assert.NoError(t, err)
	xattrs, err = sq.xattrs(plain.xattr)
	assert.NoError(t, err)
	assert.Empty(t, xattrs)
}

func TestSpecialModeBits(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{
		{Path: "setgid", Mode: fs.ModeDir | fs.ModeSetgid | 0o755},
		{Path: "setuid", Mode: fs.ModeSetuid | 0o755},
		{Path: "sticky", Mode: fs.ModeDir | fs.ModeSticky | 0o777},
	}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	defer sq.Close()
	root, err := sq.readInode(sq.sb.RootInode)
	assert.NoError(t, err)
	entries, err := sq.readDir(root)
	assert.NoError(t, err)
	var modes []uint16
	for _, e := range entries {
		in, err := sq.readInode(e.inodeRef)
		assert.NoError(t, err)
		modes = append(modes, in.mode)
	}
	assert.Equal(t, []uint16{0o2755, 0o4755, 0o1777}, modes)
}

func TestNewReaderErrors(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{{Path: "file", Mode: 0o644}}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}

	tests := map[string]struct {
		data []byte
		err  string
	}{
		"empty": {
			data: nil,
			err:  "reading squashfs superblock",
		},
		"bad magic": {
			data: append([]byte("nope"), img[4:]...),
			err:  "not a squashfs filesystem",
		},
		"lz4": {
			data: patch(img, 20, 5),
			err:  "unsupported squashfs compression: lz4",
		},
		"lzo": {
			data: patch(img, 20, 3),
			err:  "unsupported squashfs compression: lzo",
		},
		"version 3": {
			data: patch(img, 28, 3),
			err:  "unsupported squashfs version 3.0",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewReader(bytes.NewReader(tt.data))
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func patch(img []byte, offset int, value byte) []byte {
	out := append([]byte(nil), img...)
	out[offset] = value
	return out
}

func mustLstat(t *testing.T, p string) fs.FileInfo {
	fi, err := os.Lstat(p)
	assert.NoError(t, err)
	return fi
}
