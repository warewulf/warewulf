// Package squashfstest builds small SquashFS 4.0 images for tests.
package squashfstest

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io/fs"
	"math/bits"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/apptainer/sif/v2/pkg/sif"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// Compression selects the compressor used for an image.
type Compression uint16

const (
	Gzip Compression = 1
	XZ   Compression = 4
	Zstd Compression = 6
)

const (
	blockSize         = 4096
	metadataBlockSize = 8192
	invalidTable      = 0xFFFFFFFFFFFFFFFF
	noFragment        = 0xFFFFFFFF
	noXattr           = 0xFFFFFFFF
	blockUncompressed = 1 << 24
	mdUncompressed    = 1 << 15
	flagNoXattrs      = 0x0200
)

const (
	typeDir = iota + 1
	typeFile
	typeSymlink
	typeBlockDev
	typeCharDev
	typeFifo
	typeSocket
)

// extended inode types are the basic types plus 7
const extended = 7

// Entry describes a filesystem entry. Parent directories that are not
// listed are created with mode 0755.
type Entry struct {
	// Path is slash-separated and relative to the root, e.g. "bin/sh".
	Path    string
	Mode    fs.FileMode
	ModTime time.Time
	// Content is the data of a regular file.
	Content []byte
	// Target is the target of a symlink.
	Target string
	// Link, if set, makes this entry a hard link to the entry at Link.
	Link   string
	Major  uint32
	Minor  uint32
	Xattrs map[string][]byte
}

// Options configures an image. All entries are owned by root and the block
// size is 4096.
type Options struct {
	// Compression defaults to Gzip.
	Compression Compression
	// NoFragments stores file tails in data blocks instead of fragments.
	NoFragments bool
}

type node struct {
	e        Entry
	children map[string]*node
	link     *node
	nlink    uint32
	inum     uint32
	ref      uint64
	xattrIdx uint32

	blocksStart uint64
	blockSizes  []uint32
	fragIdx     uint32
	fragOffset  uint32

	dirBlock  uint32
	dirOffset uint16
	dirSize   uint32
}

func (n *node) basicType() uint16 {
	m := n.e.Mode
	switch {
	case m.IsDir():
		return typeDir
	case m&fs.ModeSymlink != 0:
		return typeSymlink
	case m&fs.ModeDevice != 0 && m&fs.ModeCharDevice != 0:
		return typeCharDev
	case m&fs.ModeDevice != 0:
		return typeBlockDev
	case m&fs.ModeNamedPipe != 0:
		return typeFifo
	case m&fs.ModeSocket != 0:
		return typeSocket
	}
	return typeFile
}

func (n *node) resolve() *node {
	if n.link != nil {
		return n.link
	}
	return n
}

func (n *node) sortedChildren() []string {
	names := make([]string, 0, len(n.children))
	for name := range n.children {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type builder struct {
	opts      Options
	img       []byte
	zstd      *zstd.Encoder
	inodes    *mdWriter
	dirs      *mdWriter
	xattrKV   *mdWriter
	xattrIDs  []byte
	xattrs    uint32
	frag      []byte
	frags     []byte
	fragCount uint32
	inodeNum  uint32
}

// Build returns a SquashFS image containing entries.
func Build(entries []Entry, opts Options) ([]byte, error) {
	if opts.Compression == 0 {
		opts.Compression = Gzip
	}
	b := &builder{
		opts: opts,
		img:  make([]byte, 96),
	}
	if opts.Compression == Zstd {
		enc, err := zstd.NewWriter(nil)
		if err != nil {
			return nil, err
		}
		defer enc.Close()
		b.zstd = enc
	}
	b.inodes = &mdWriter{b: b}
	b.dirs = &mdWriter{b: b}
	b.xattrKV = &mdWriter{b: b}

	root, err := buildTree(entries)
	if err != nil {
		return nil, err
	}
	b.number(root)
	if err := b.writeData(root); err != nil {
		return nil, err
	}
	if err := b.flushFragment(); err != nil {
		return nil, err
	}
	b.writeInodes(root)
	b.writeDirs(root)

	var sb struct {
		Magic             uint32
		InodeCount        uint32
		ModTime           uint32
		BlockSize         uint32
		FragCount         uint32
		Compression       uint16
		BlockLog          uint16
		Flags             uint16
		IDCount           uint16
		VersionMajor      uint16
		VersionMinor      uint16
		RootInode         uint64
		BytesUsed         uint64
		IDTableStart      uint64
		XattrIDTableStart uint64
		InodeTableStart   uint64
		DirTableStart     uint64
		FragTableStart    uint64
		ExportTableStart  uint64
	}
	sb.Magic = 0x73717368
	sb.InodeCount = b.inodeNum
	sb.BlockSize = blockSize
	sb.Compression = uint16(opts.Compression)
	sb.BlockLog = uint16(bits.TrailingZeros32(blockSize))
	sb.VersionMajor = 4
	sb.RootInode = root.ref
	sb.ExportTableStart = invalidTable

	sb.InodeTableStart = uint64(len(b.img))
	b.img = append(b.img, b.inodes.finish()...)
	sb.DirTableStart = uint64(len(b.img))
	b.img = append(b.img, b.dirs.finish()...)

	sb.FragTableStart = invalidTable
	if b.fragCount > 0 {
		sb.FragCount = b.fragCount
		sb.FragTableStart = b.writeLookupTable(b.frags)
	}

	// the only id is root, at index 0
	sb.IDCount = 1
	sb.IDTableStart = b.writeLookupTable(le(uint32(0)))

	sb.XattrIDTableStart = invalidTable
	if b.xattrs == 0 {
		sb.Flags |= flagNoXattrs
	} else {
		kvStart := uint64(len(b.img))
		b.img = append(b.img, b.xattrKV.finish()...)
		var locs []uint64
		for i := 0; i < len(b.xattrIDs); i += metadataBlockSize {
			locs = append(locs, uint64(len(b.img)))
			b.img = append(b.img, b.metadataBlock(b.xattrIDs[i:min(i+metadataBlockSize, len(b.xattrIDs))])...)
		}
		sb.XattrIDTableStart = uint64(len(b.img))
		b.img = append(b.img, le(kvStart, b.xattrs, uint32(0), locs)...)
	}

	sb.BytesUsed = uint64(len(b.img))
	copy(b.img, le(sb))
	return b.img, nil
}

func buildTree(entries []Entry) (*node, error) {
	root := &node{e: Entry{Mode: fs.ModeDir | 0o755}, children: map[string]*node{}}
	byPath := map[string]*node{}
	lookup := func(p string) *node {
		n := root
		for _, part := range strings.Split(p, "/") {
			child, ok := n.children[part]
			if !ok {
				child = &node{e: Entry{Path: part, Mode: fs.ModeDir | 0o755}, children: map[string]*node{}}
				n.children[part] = child
			}
			n = child
		}
		return n
	}
	for _, e := range entries {
		p := path.Clean(e.Path)
		if p == "." || strings.HasPrefix(p, "/") || strings.HasPrefix(p, "..") {
			return nil, fmt.Errorf("invalid path %q", e.Path)
		}
		parent := root
		if dir := path.Dir(p); dir != "." {
			parent = lookup(dir)
		}
		n := &node{e: e}
		if e.Mode.IsDir() {
			n.children = map[string]*node{}
			if existing, ok := parent.children[path.Base(p)]; ok {
				n.children = existing.children
			}
		}
		parent.children[path.Base(p)] = n
		byPath[p] = n
	}
	for _, n := range byPath {
		if n.e.Link == "" {
			continue
		}
		target, ok := byPath[path.Clean(n.e.Link)]
		if !ok || target.e.Link != "" || target.e.Mode.IsDir() {
			return nil, fmt.Errorf("invalid hard link target %q", n.e.Link)
		}
		n.link = target
	}
	return root, nil
}

// number assigns inode numbers and link counts.
func (b *builder) number(n *node) {
	b.inodeNum++
	n.inum = b.inodeNum
	n.xattrIdx = noXattr
	n.nlink = 1
	if n.children != nil {
		n.nlink = 2
	}
	for _, name := range n.sortedChildren() {
		c := n.children[name]
		if c.link != nil {
			continue
		}
		b.number(c)
		if c.children != nil {
			n.nlink++
		}
	}
	for _, name := range n.sortedChildren() {
		if c := n.children[name]; c.link != nil {
			c.link.nlink++
		}
	}
}

func (b *builder) writeData(n *node) error {
	for _, name := range n.sortedChildren() {
		c := n.children[name]
		if c.link != nil {
			continue
		}
		b.addXattrs(c)
		if c.children != nil {
			if err := b.writeData(c); err != nil {
				return err
			}
			continue
		}
		if c.basicType() != typeFile {
			continue
		}
		if err := b.writeFile(c); err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) writeFile(n *node) error {
	data := n.e.Content
	bs := blockSize
	n.blocksStart = uint64(len(b.img))
	n.fragIdx = noFragment
	full := len(data) / bs
	tail := data[full*bs:]
	useFragment := len(tail) > 0 && !b.opts.NoFragments
	count := full
	if len(tail) > 0 && !useFragment {
		count++
	}
	for i := 0; i < count; i++ {
		blk := data[i*bs : min((i+1)*bs, len(data))]
		if bytes.Count(blk, []byte{0}) == len(blk) {
			// sparse
			n.blockSizes = append(n.blockSizes, 0)
			continue
		}
		size, err := b.writeBlock(blk)
		if err != nil {
			return err
		}
		n.blockSizes = append(n.blockSizes, size)
	}
	if useFragment {
		if len(b.frag)+len(tail) > bs {
			if err := b.flushFragment(); err != nil {
				return err
			}
		}
		n.fragIdx = b.fragCount
		n.fragOffset = uint32(len(b.frag))
		b.frag = append(b.frag, tail...)
	}
	return nil
}

func (b *builder) writeBlock(blk []byte) (uint32, error) {
	c, err := b.compress(blk)
	if err != nil {
		return 0, err
	}
	if len(c) < len(blk) {
		b.img = append(b.img, c...)
		return uint32(len(c)), nil
	}
	b.img = append(b.img, blk...)
	return uint32(len(blk)) | blockUncompressed, nil
}

func (b *builder) flushFragment() error {
	if len(b.frag) == 0 {
		return nil
	}
	start := uint64(len(b.img))
	size, err := b.writeBlock(b.frag)
	if err != nil {
		return err
	}
	b.frags = append(b.frags, le(start, size, uint32(0))...)
	b.fragCount++
	b.frag = nil
	return nil
}

func (b *builder) addXattrs(n *node) {
	n.xattrIdx = noXattr
	if len(n.e.Xattrs) == 0 {
		return
	}
	names := make([]string, 0, len(n.e.Xattrs))
	for name := range n.e.Xattrs {
		names = append(names, name)
	}
	sort.Strings(names)
	ref := b.xattrKV.ref()
	size, count := 0, 0
	for _, name := range names {
		var typ uint16
		var suffix string
		switch {
		case strings.HasPrefix(name, "user."):
			typ, suffix = 0, strings.TrimPrefix(name, "user.")
		case strings.HasPrefix(name, "trusted."):
			typ, suffix = 1, strings.TrimPrefix(name, "trusted.")
		case strings.HasPrefix(name, "security."):
			typ, suffix = 2, strings.TrimPrefix(name, "security.")
		default:
			continue
		}
		value := n.e.Xattrs[name]
		kv := append(le(typ, uint16(len(suffix))), suffix...)
		kv = append(kv, le(uint32(len(value)))...)
		kv = append(kv, value...)
		b.xattrKV.write(kv)
		size += len(kv)
		count++
	}
	n.xattrIdx = b.xattrs
	b.xattrs++
	b.xattrIDs = append(b.xattrIDs, le(ref, uint32(count), uint32(size))...)
}

// writeInodes writes the inodes of all non-directory entries.
func (b *builder) writeInodes(n *node) {
	for _, name := range n.sortedChildren() {
		c := n.children[name]
		if c.link != nil {
			continue
		}
		if c.children != nil {
			b.writeInodes(c)
		} else {
			b.writeInode(c)
		}
	}
}

// writeDirs writes directory listings and inodes, children first.
func (b *builder) writeDirs(n *node) {
	for _, name := range n.sortedChildren() {
		if c := n.children[name]; c.link == nil && c.children != nil {
			b.writeDirs(c)
		}
	}
	b.writeListing(n)
	b.writeInode(n)
}

func (b *builder) writeListing(n *node) {
	start := b.dirs.ref()
	names := n.sortedChildren()
	written := 0
	for i := 0; i < len(names); {
		first := n.children[names[i]].resolve()
		block := uint32(first.ref >> 16)
		j := i
		for j < len(names) && j-i < 256 && uint32(n.children[names[j]].resolve().ref>>16) == block {
			j++
		}
		hdr := le(uint32(j-i-1), block, first.inum)
		b.dirs.write(hdr)
		written += len(hdr)
		for k := i; k < j; k++ {
			c := n.children[names[k]].resolve()
			e := append(le(uint16(c.ref), int16(c.inum-first.inum), c.basicType(), uint16(len(names[k])-1)), names[k]...)
			b.dirs.write(e)
			written += len(e)
		}
		i = j
	}
	n.dirBlock = uint32(start >> 16)
	n.dirOffset = uint16(start)
	n.dirSize = uint32(written + 3)
}

func (b *builder) writeInode(n *node) {
	e := n.e
	typ := n.basicType()
	hasXattr := n.xattrIdx != noXattr
	if hasXattr || (typ == typeFile && n.nlink > 1) {
		typ += extended
	}
	var mtime uint32
	if !e.ModTime.IsZero() {
		mtime = uint32(e.ModTime.Unix())
	}
	perm := uint16(e.Mode.Perm())
	if e.Mode&fs.ModeSetuid != 0 {
		perm |= 0o4000
	}
	if e.Mode&fs.ModeSetgid != 0 {
		perm |= 0o2000
	}
	if e.Mode&fs.ModeSticky != 0 {
		perm |= 0o1000
	}
	buf := le(typ, perm, uint16(0), uint16(0), mtime, n.inum)

	rdev := (e.Minor & 0xff) | (e.Major << 8) | ((e.Minor &^ 0xff) << 12)
	size := uint64(len(e.Content))
	switch typ {
	case typeDir:
		buf = append(buf, le(n.dirBlock, n.nlink, uint16(n.dirSize), n.dirOffset, uint32(0))...)
	case typeDir + extended:
		buf = append(buf, le(n.nlink, n.dirSize, n.dirBlock, uint32(0), uint16(0), n.dirOffset, n.xattrIdx)...)
	case typeFile:
		buf = append(buf, le(uint32(n.blocksStart), n.fragIdx, n.fragOffset, uint32(size), n.blockSizes)...)
	case typeFile + extended:
		buf = append(buf, le(n.blocksStart, size, uint64(0), n.nlink, n.fragIdx, n.fragOffset, n.xattrIdx, n.blockSizes)...)
	case typeSymlink:
		buf = append(append(buf, le(n.nlink, uint32(len(e.Target)))...), e.Target...)
	case typeSymlink + extended:
		buf = append(append(buf, le(n.nlink, uint32(len(e.Target)))...), e.Target...)
		buf = append(buf, le(n.xattrIdx)...)
	case typeBlockDev, typeCharDev:
		buf = append(buf, le(n.nlink, rdev)...)
	case typeBlockDev + extended, typeCharDev + extended:
		buf = append(buf, le(n.nlink, rdev, n.xattrIdx)...)
	case typeFifo, typeSocket:
		buf = append(buf, le(n.nlink)...)
	case typeFifo + extended, typeSocket + extended:
		buf = append(buf, le(n.nlink, n.xattrIdx)...)
	}
	n.ref = b.inodes.ref()
	b.inodes.write(buf)
}

func (b *builder) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	switch b.opts.Compression {
	case Gzip:
		w := zlib.NewWriter(&buf)
		if _, err := w.Write(data); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
	case XZ:
		w, err := xz.NewWriter(&buf)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(data); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
	case Zstd:
		return b.zstd.EncodeAll(data, nil), nil
	default:
		return nil, fmt.Errorf("unsupported compression %d", b.opts.Compression)
	}
	return buf.Bytes(), nil
}

// metadataBlock encodes data as a single metadata block.
func (b *builder) metadataBlock(data []byte) []byte {
	c, err := b.compress(data)
	if err != nil || len(c) >= len(data) {
		return append(le(uint16(len(data))|mdUncompressed), data...)
	}
	return append(le(uint16(len(c))), c...)
}

func (b *builder) writeLookupTable(data []byte) uint64 {
	var locs []uint64
	for i := 0; i < len(data); i += metadataBlockSize {
		locs = append(locs, uint64(len(b.img)))
		b.img = append(b.img, b.metadataBlock(data[i:min(i+metadataBlockSize, len(data))])...)
	}
	start := uint64(len(b.img))
	b.img = append(b.img, le(locs)...)
	return start
}

// mdWriter writes a table of metadata blocks.
type mdWriter struct {
	b       *builder
	out     []byte
	pending []byte
}

// ref returns a reference to the current write position.
func (w *mdWriter) ref() uint64 {
	return uint64(len(w.out))<<16 | uint64(len(w.pending))
}

func (w *mdWriter) write(p []byte) {
	w.pending = append(w.pending, p...)
	for len(w.pending) >= metadataBlockSize {
		w.out = append(w.out, w.b.metadataBlock(w.pending[:metadataBlockSize])...)
		w.pending = append([]byte(nil), w.pending[metadataBlockSize:]...)
	}
}

func (w *mdWriter) finish() []byte {
	if len(w.pending) > 0 {
		w.out = append(w.out, w.b.metadataBlock(w.pending)...)
		w.pending = nil
	}
	return w.out
}

// WriteSIF writes a SIF image to path whose primary system partition is a
// SquashFS image containing entries.
func WriteSIF(path string, entries []Entry) error {
	img, err := Build(entries, Options{})
	if err != nil {
		return err
	}
	di, err := sif.NewDescriptorInput(sif.DataPartition, bytes.NewReader(img),
		sif.OptPartitionMetadata(sif.FsSquash, sif.PartPrimSys, runtime.GOARCH))
	if err != nil {
		return err
	}
	f, err := sif.CreateContainerAtPath(path, sif.OptCreateWithDescriptors(di))
	if err != nil {
		return err
	}
	return f.UnloadContainer()
}

// le encodes values in little-endian byte order.
func le(values ...any) []byte {
	var buf bytes.Buffer
	for _, v := range values {
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			panic(err)
		}
	}
	return buf.Bytes()
}
