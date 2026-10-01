package squashfs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

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
		// a hard link in a directory that sorts before its target's
		{Path: "bin/sudo", Link: "usr/bin/sudo"},
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
	assert.NoError(t, sq.Extract(dst))

	content, err := os.ReadFile(filepath.Join(dst, "bin/sh"))
	assert.NoError(t, err)
	assert.Equal(t, "shell", string(content))
	assert.DirExists(t, filepath.Join(dst, "etc"))
	assert.FileExists(t, filepath.Join(dst, "kept"))
}

func TestExtractDoesNotFollowExistingSymlinks(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{
		{Path: "etc/passwd", Mode: 0o644, Content: []byte("root")},
		{Path: "lib", Mode: 0o644, Content: []byte("lib")},
		{Path: "var/x/y", Mode: 0o644},
	}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	outside := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(outside, "file"), []byte("outside"), 0o644))
	dst := t.TempDir()
	assert.NoError(t, os.Symlink(outside, filepath.Join(dst, "etc")))
	assert.NoError(t, os.Symlink(filepath.Join(outside, "file"), filepath.Join(dst, "lib")))
	assert.NoError(t, os.Mkdir(filepath.Join(dst, "var"), 0o755))
	assert.NoError(t, os.Symlink(outside, filepath.Join(dst, "var/x")))

	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	assert.NoError(t, sq.Extract(dst))

	assert.DirExists(t, filepath.Join(dst, "etc"))
	assert.FileExists(t, filepath.Join(dst, "etc/passwd"))
	assert.FileExists(t, filepath.Join(dst, "var/x/y"))
	written, err := os.ReadDir(outside)
	assert.NoError(t, err)
	assert.Len(t, written, 1)
	content, err := os.ReadFile(filepath.Join(outside, "file"))
	assert.NoError(t, err)
	assert.Equal(t, "outside", string(content))
}

func TestExtractXattrs(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{
		{Path: "file", Mode: 0o644, Xattrs: map[string][]byte{
			"user.one":         []byte("1"),
			"trusted.two":      []byte("2"),
			"security.selinux": []byte("system_u:object_r:image_t:s0"),
		}},
	}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	dst := t.TempDir()
	if !assert.NoError(t, sq.Extract(dst)) {
		return
	}
	p := filepath.Join(dst, "file")
	buf := make([]byte, 64)

	n, err := unix.Lgetxattr(p, "user.one", buf)
	if errors.Is(err, unix.ENOTSUP) {
		t.Skip("extended attributes are not supported on the temporary directory")
	}
	if assert.NoError(t, err) {
		assert.Equal(t, "1", string(buf[:n]))
	}

	n, err = unix.Lgetxattr(p, "trusted.two", buf)
	if os.Geteuid() == 0 {
		if assert.NoError(t, err) {
			assert.Equal(t, "2", string(buf[:n]))
		}
	} else {
		assert.Error(t, err)
	}

	// The host may label the file itself, but never with the image's label.
	if n, err = unix.Lgetxattr(p, "security.selinux", buf); err == nil {
		assert.NotEqual(t, "system_u:object_r:image_t:s0", string(buf[:n]))
	}
}

func TestXattrs(t *testing.T) {
	want := map[string][]byte{
		"user.one":            []byte("1"),
		"security.capability": {1, 2, 3},
		"trusted.three":       []byte("three"),
	}
	img, err := squashfstest.Build([]squashfstest.Entry{
		// sorts after copy, so its values are stored out of line
		{Path: "file", Mode: 0o644, Xattrs: want},
		{Path: "copy", Mode: 0o644, Xattrs: want},
		{Path: "plain", Mode: 0o644},
	}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	root, err := sq.readInode(sq.sb.RootInode)
	assert.NoError(t, err)
	entries, err := sq.readDir(root)
	assert.NoError(t, err)
	assert.Len(t, entries, 3)

	for _, e := range entries[:2] {
		file, err := sq.readInode(e.inodeRef)
		assert.NoError(t, err)
		xattrs, err := sq.xattrs(file.xattr)
		assert.NoError(t, err)
		assert.Equal(t, want, xattrs, e.name)
	}

	plain, err := sq.readInode(entries[2].inodeRef)
	assert.NoError(t, err)
	xattrs, err := sq.xattrs(plain.xattr)
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
		"truncated": {
			data: img[:len(img)-1],
			err:  "squashfs image is truncated",
		},
		"fragment count larger than the image": {
			data: patch(img, 19, 0x10),
			err:  "invalid table entry count",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewReader(bytes.NewReader(tt.data))
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestNewReaderKeepsReadError(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{{Path: "file", Mode: 0o644}}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	errRead := errors.New("read failed")
	_, err = NewReader(failingReaderAt{img: img[:superblockSize], err: errRead})
	assert.ErrorIs(t, err, errRead)
}

// failingReaderAt returns img and then fails reads past its end with err.
type failingReaderAt struct {
	img []byte
	err error
}

func (r failingReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off+int64(len(p)) > int64(len(r.img)) {
		return 0, r.err
	}
	return copy(p, r.img[off:]), nil
}

func TestExtractRejectsMalformedDirectories(t *testing.T) {
	tests := map[string]struct {
		edit func(t *testing.T, sq *Reader)
		err  string
	}{
		"duplicate name": {
			edit: func(t *testing.T, sq *Reader) {
				patchEntry(t, sq, "dir2", func(e []byte) { e[11] = '1' })
			},
			err: "directory entries are not sorted",
		},
		"shared listing": {
			edit: func(t *testing.T, sq *Reader) {
				var offset []byte
				patchEntry(t, sq, "dir1", func(e []byte) { offset = bytes.Clone(e[:2]) })
				patchEntry(t, sq, "dir2", func(e []byte) { copy(e, offset) })
			},
			err: "directory is referenced more than once",
		},
		"cycle": {
			edit: func(t *testing.T, sq *Reader) {
				patchEntry(t, sq, "dir1", func(e []byte) { binary.LittleEndian.PutUint16(e, uint16(sq.sb.RootInode)) })
			},
			err: "directory is referenced more than once",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			img, err := squashfstest.Build([]squashfstest.Entry{
				{Path: "dir1/file", Mode: 0o644},
				{Path: "dir2/file", Mode: 0o644},
			}, squashfstest.Options{})
			if !assert.NoError(t, err) {
				return
			}
			sq, err := NewReader(bytes.NewReader(img))
			if !assert.NoError(t, err) {
				return
			}
			tt.edit(t, sq)
			assert.ErrorContains(t, sq.Extract(t.TempDir()), tt.err)
		})
	}
}

func TestXZBlock(t *testing.T) {
	// a stream header followed by a 12-byte block header with one filter
	block := func(flags, filter, dict byte) []byte {
		return append([]byte("\xfd7zXZ\x00\x00\x01\x00\x00\x00\x00"),
			2, flags, filter, 1, dict, 0, 0, 0, 0, 0, 0, 0, 'x')
	}
	tests := map[string]struct {
		src     []byte
		dictCap int
		err     string
	}{
		"4 KiB dictionary":   {src: block(0, 0x21, 0), dictCap: 4 << 10},
		"1 MiB dictionary":   {src: block(0, 0x21, 16), dictCap: 1 << 20},
		"2 MiB dictionary":   {src: block(0, 0x21, 18), err: "xz dictionary size 2097152 exceeds"},
		"BCJ filter":         {src: block(1, 0x04, 0), err: "such as BCJ"},
		"other filter":       {src: block(0, 0x03, 0), err: "supported LZMA2 filter"},
		"reserved flag bits": {src: block(0x04, 0x21, 16), err: "invalid xz block header"},
		"truncated":          {src: []byte("\xfd7zXZ\x00"), err: "invalid xz stream header"},
		"sizes in the header": {
			// as written by liblzma for mksquashfs: compressed size 129
			// and uncompressed size 5 precede the filter
			src: append([]byte("\xfd7zXZ\x00\x00\x01\x00\x00\x00\x00"),
				2, 0xc0, 0x81, 0x01, 0x05, 0x21, 1, 16, 0, 0, 0, 0, 'x'),
			dictCap: 1 << 20,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			data, dictCap, err := xzBlock(tt.src)
			if tt.err != "" {
				assert.ErrorContains(t, err, tt.err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.dictCap, dictCap)
			assert.Equal(t, []byte("x"), data)
		})
	}
}

func TestXattrLimits(t *testing.T) {
	value := bytes.Repeat([]byte("x"), maxXattrValueSize)
	many := map[string][]byte{}
	for i := 0; i*maxXattrValueSize <= maxXattrSize; i++ {
		many[fmt.Sprintf("user.%02d", i)] = value
	}
	tests := map[string]struct {
		xattrs map[string][]byte
		err    string
	}{
		"value too large": {xattrs: map[string][]byte{"user.big": append(value, 'x')}, err: "is too large"},
		"total too large": {xattrs: many, err: "extended attributes exceed"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			img, err := squashfstest.Build([]squashfstest.Entry{{Path: "file", Mode: 0o644, Xattrs: tt.xattrs}}, squashfstest.Options{})
			if !assert.NoError(t, err) {
				return
			}
			sq, err := NewReader(bytes.NewReader(img))
			if !assert.NoError(t, err) {
				return
			}
			assert.ErrorContains(t, sq.Extract(t.TempDir()), tt.err)
		})
	}
}

// TestBoundedAllocations checks that counts taken from the image do not size
// allocations before the data behind them has been read.
func TestBoundedAllocations(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{
		{Path: "file", Mode: 0o644, Xattrs: map[string][]byte{"user.one": []byte("1")}},
	}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	const limit = 8 << 20

	t.Run("xattr count", func(t *testing.T) {
		sq.xattrIDs[0].Count = 1 << 20
		assert.Less(t, allocated(func() { _, _ = sq.xattrs(0) }), uint64(limit))
	})

	t.Run("block count", func(t *testing.T) {
		m, err := sq.newMetadataReader(int64(sq.sb.InodeTableStart), 0, 0)
		if !assert.NoError(t, err) {
			return
		}
		in := &inode{fileSize: uint64(maxBlockCount) * uint64(sq.sb.BlockSize), fragIdx: noFragment}
		assert.Less(t, allocated(func() { assert.Error(t, sq.readBlockSizes(m, in)) }), uint64(limit))
	})

	t.Run("lookup table", func(t *testing.T) {
		// One metadata block, named by every location in the index that
		// follows it, in an image whose declared size is mostly a hole.
		count := uint64(maxTableSize/16 + 1)
		blocks := (count*16 + metadataBlockSize - 1) / metadataBlockSize
		data := binary.LittleEndian.AppendUint16(nil, metadataBlockSize|mdUncompressed)
		data = append(data, make([]byte, metadataBlockSize+blocks*8)...)
		sparse := &Reader{r: bytes.NewReader(data), sb: superblock{BytesUsed: 1 << 40}, metadata: map[int64]metadataBlock{}}
		var err error
		assert.Less(t, allocated(func() {
			_, err = readLookupTable[fragmentEntry](sparse, metadataBlockSize+2, count)
		}), uint64(limit))
		assert.ErrorContains(t, err, "invalid table entry count")
	})
}

func TestMetadataCacheIsBounded(t *testing.T) {
	// Every even position decodes as an uncompressed 8 KiB metadata block.
	data := bytes.Repeat([]byte{0x00, 0xa0}, maxCachedBlocks+metadataBlockSize)
	sq := &Reader{r: bytes.NewReader(data), sb: superblock{BytesUsed: uint64(len(data))}, metadata: map[int64]metadataBlock{}}
	for pos := int64(0); pos < 2*(maxCachedBlocks+10); pos += 2 {
		_, _, err := sq.readMetadataBlock(pos)
		if !assert.NoError(t, err) {
			return
		}
	}
	assert.LessOrEqual(t, len(sq.metadata), maxCachedBlocks)
}

func TestBlockSizesAreBounded(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{{Path: "file", Mode: 0o644, Content: []byte("x")}}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	// sizes with bits above the block size set, which an image may declare
	huge := uint32(1<<27 | blockUncompressed)

	t.Run("data block", func(t *testing.T) {
		f, err := os.Create(filepath.Join(t.TempDir(), "file"))
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			_ = f.Close()
		}()
		in := &inode{fileSize: uint64(sq.sb.BlockSize), fragIdx: noFragment, blockSizes: []uint32{huge}}
		assert.ErrorContains(t, sq.writeFile(in, f), "invalid size")
	})

	t.Run("fragment", func(t *testing.T) {
		sq.frags[0].Size = huge
		sq.fragIdx = noFragment
		_, err := sq.fragment(0)
		assert.ErrorContains(t, err, "invalid size")
	})
}

func TestOwnership(t *testing.T) {
	img, err := squashfstest.Build([]squashfstest.Entry{
		{Path: "mine", Mode: 0o644, UID: 1000, GID: 2000},
		{Path: "root", Mode: 0o644},
		{Path: "theirs", Mode: 0o644, UID: 2000, GID: 1000},
	}, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	root, err := sq.readInode(sq.sb.RootInode)
	assert.NoError(t, err)
	entries, err := sq.readDir(root)
	assert.NoError(t, err)
	var owners [][2]uint32
	for _, e := range entries {
		in, err := sq.readInode(e.inodeRef)
		assert.NoError(t, err)
		owners = append(owners, [2]uint32{in.uid, in.gid})
	}
	assert.Equal(t, [][2]uint32{{1000, 2000}, {0, 0}, {2000, 1000}}, owners)
}

func TestLargeDirectory(t *testing.T) {
	// a listing larger than the 16-bit size of a basic directory inode
	var entries []squashfstest.Entry
	for i := 0; i < 2200; i++ {
		entries = append(entries, squashfstest.Entry{Path: fmt.Sprintf("big/file-with-a-long-name-%04d", i), Mode: 0o644})
	}
	img, err := squashfstest.Build(entries, squashfstest.Options{})
	if !assert.NoError(t, err) {
		return
	}
	sq, err := NewReader(bytes.NewReader(img))
	if !assert.NoError(t, err) {
		return
	}
	root, err := sq.readInode(sq.sb.RootInode)
	assert.NoError(t, err)
	top, err := sq.readDir(root)
	if !assert.NoError(t, err) || !assert.Len(t, top, 1) {
		return
	}
	big, err := sq.readInode(top[0].inodeRef)
	assert.NoError(t, err)
	listing, err := sq.readDir(big)
	assert.NoError(t, err)
	assert.Len(t, listing, len(entries))
}

// allocated returns the number of bytes allocated while running f.
func allocated(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// patchEntry edits the directory entry called name in the first block of the
// directory table. An entry holds the inode offset, the inode number delta,
// the type, the name size minus one, and the name.
func patchEntry(t *testing.T, sq *Reader, name string, edit func(entry []byte)) {
	t.Helper()
	pos := int64(sq.sb.DirTableStart)
	data, next, err := sq.readMetadataBlock(pos)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	data = bytes.Clone(data)
	i := bytes.Index(data, append(binary.LittleEndian.AppendUint16(nil, uint16(len(name)-1)), name...))
	if !assert.GreaterOrEqual(t, i, 6, name) {
		t.FailNow()
	}
	edit(data[i-6:])
	sq.metadata[pos] = metadataBlock{data: data, next: next}
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
