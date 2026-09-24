// Package squashfs reads SquashFS 4.0 filesystem images.
//
// It implements only what is needed to extract a complete image to a
// directory: gzip, xz, and zstd compression, fragments, sparse files, hard
// links, device nodes, and extended attributes. The on-disk format is
// described at https://dr-emann.github.io/squashfs/.
package squashfs

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

const (
	superblockMagic   = 0x73717368
	superblockSize    = 96
	metadataBlockSize = 8192
	maxBlockSize      = 1 << 20
	maxBlockCount     = 1 << 24
	maxXattrValueSize = 1 << 20

	invalidTable      = 0xFFFFFFFFFFFFFFFF
	noFragment        = 0xFFFFFFFF
	noXattr           = 0xFFFFFFFF
	blockUncompressed = 1 << 24
	mdUncompressed    = 1 << 15
)

const (
	compGzip = 1
	compLZMA = 2
	compLZO  = 3
	compXZ   = 4
	compLZ4  = 5
	compZstd = 6
)

var compressionNames = map[uint16]string{
	compGzip: "gzip",
	compLZMA: "lzma",
	compLZO:  "lzo",
	compXZ:   "xz",
	compLZ4:  "lz4",
	compZstd: "zstd",
}

const (
	typeDir = iota + 1
	typeFile
	typeSymlink
	typeBlockDev
	typeCharDev
	typeFifo
	typeSocket
	typeExtDir
	typeExtFile
	typeExtSymlink
	typeExtBlockDev
	typeExtCharDev
	typeExtFifo
	typeExtSocket
)

type superblock struct {
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

type metadataBlock struct {
	data []byte
	next int64
}

type fragmentEntry struct {
	start uint64
	size  uint32
}

type xattrID struct {
	ref   uint64
	count uint32
}

// Reader reads a SquashFS image.
type Reader struct {
	r          io.ReaderAt
	sb         superblock
	zstd       *zstd.Decoder
	ids        []uint32
	frags      []fragmentEntry
	xattrIDs   []xattrID
	xattrStart int64
	metadata   map[int64]metadataBlock
	fragIdx    uint32
	fragData   []byte
}

// NewReader returns a Reader for the SquashFS image in r. The caller must
// call Close when done.
func NewReader(r io.ReaderAt) (*Reader, error) {
	sq := &Reader{
		r:        r,
		metadata: map[int64]metadataBlock{},
		fragIdx:  noFragment,
	}
	if err := binary.Read(io.NewSectionReader(r, 0, superblockSize), binary.LittleEndian, &sq.sb); err != nil {
		return nil, fmt.Errorf("reading squashfs superblock: %w", err)
	}
	if sq.sb.Magic != superblockMagic {
		return nil, errors.New("not a squashfs filesystem")
	}
	if sq.sb.VersionMajor != 4 || sq.sb.VersionMinor != 0 {
		return nil, fmt.Errorf("unsupported squashfs version %d.%d", sq.sb.VersionMajor, sq.sb.VersionMinor)
	}
	if sq.sb.BlockSize == 0 || sq.sb.BlockSize > maxBlockSize || sq.sb.BlockLog > 20 || sq.sb.BlockSize != 1<<sq.sb.BlockLog {
		return nil, fmt.Errorf("invalid squashfs block size %d", sq.sb.BlockSize)
	}
	switch sq.sb.Compression {
	case compGzip, compXZ:
	case compZstd:
		d, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(maxBlockSize))
		if err != nil {
			return nil, err
		}
		sq.zstd = d
	default:
		name, ok := compressionNames[sq.sb.Compression]
		if !ok {
			name = fmt.Sprintf("id %d", sq.sb.Compression)
		}
		return nil, fmt.Errorf("unsupported squashfs compression: %s", name)
	}
	if err := sq.readTables(); err != nil {
		sq.Close()
		return nil, err
	}
	return sq, nil
}

// Close releases resources held by the Reader.
func (sq *Reader) Close() {
	if sq.zstd != nil {
		sq.zstd.Close()
	}
}

func (sq *Reader) readTables() error {
	data, err := sq.readLookupTable(sq.sb.IDTableStart, uint64(sq.sb.IDCount), 4)
	if err != nil {
		return fmt.Errorf("reading id table: %w", err)
	}
	sq.ids = make([]uint32, sq.sb.IDCount)
	for i := range sq.ids {
		sq.ids[i] = binary.LittleEndian.Uint32(data[i*4:])
	}

	if sq.sb.FragCount > 0 {
		data, err := sq.readLookupTable(sq.sb.FragTableStart, uint64(sq.sb.FragCount), 16)
		if err != nil {
			return fmt.Errorf("reading fragment table: %w", err)
		}
		sq.frags = make([]fragmentEntry, sq.sb.FragCount)
		for i := range sq.frags {
			sq.frags[i].start = binary.LittleEndian.Uint64(data[i*16:])
			sq.frags[i].size = binary.LittleEndian.Uint32(data[i*16+8:])
		}
	}

	if sq.sb.XattrIDTableStart != invalidTable {
		hdr, err := sq.readAt(16, int64(sq.sb.XattrIDTableStart))
		if err != nil {
			return fmt.Errorf("reading xattr id table: %w", err)
		}
		sq.xattrStart = int64(binary.LittleEndian.Uint64(hdr))
		count := binary.LittleEndian.Uint32(hdr[8:])
		data, err := sq.readLookupTable(sq.sb.XattrIDTableStart+16, uint64(count), 16)
		if err != nil {
			return fmt.Errorf("reading xattr id table: %w", err)
		}
		sq.xattrIDs = make([]xattrID, count)
		for i := range sq.xattrIDs {
			sq.xattrIDs[i].ref = binary.LittleEndian.Uint64(data[i*16:])
			sq.xattrIDs[i].count = binary.LittleEndian.Uint32(data[i*16+8:])
		}
	}
	return nil
}

// readAt reads size bytes at off, which must lie within the filesystem.
func (sq *Reader) readAt(size, off int64) ([]byte, error) {
	if off < 0 || size < 0 || off+size > int64(sq.sb.BytesUsed) {
		return nil, fmt.Errorf("read of %d bytes at offset %d is outside the filesystem", size, off)
	}
	buf := make([]byte, size)
	if n, err := sq.r.ReadAt(buf, off); n < len(buf) {
		return nil, err
	}
	return buf, nil
}

// decompress decompresses a block that is expected to expand to no more than
// limit bytes.
func (sq *Reader) decompress(src []byte, limit int) ([]byte, error) {
	var rd io.Reader
	switch sq.sb.Compression {
	case compGzip:
		zr, err := zlib.NewReader(bytes.NewReader(src))
		if err != nil {
			return nil, err
		}

		defer func() {
			_ = zr.Close()
		}()

		rd = zr
	case compXZ:
		xr, err := xz.NewReader(bytes.NewReader(src))
		if err != nil {
			return nil, err
		}
		rd = xr
	case compZstd:
		out, err := sq.zstd.DecodeAll(src, nil)
		if err != nil {
			return nil, err
		}
		if len(out) > limit {
			return nil, fmt.Errorf("decompressed block exceeds %d bytes", limit)
		}
		return out, nil
	}
	out, err := io.ReadAll(io.LimitReader(rd, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(out) > limit {
		return nil, fmt.Errorf("decompressed block exceeds %d bytes", limit)
	}
	return out, nil
}

// readMetadataBlock returns the uncompressed contents of the metadata block
// at pos and the position of the block that follows it.
func (sq *Reader) readMetadataBlock(pos int64) ([]byte, int64, error) {
	if b, ok := sq.metadata[pos]; ok {
		return b.data, b.next, nil
	}
	hdr, err := sq.readAt(2, pos)
	if err != nil {
		return nil, 0, err
	}
	h := binary.LittleEndian.Uint16(hdr)
	size := int64(h &^ mdUncompressed)
	if size > metadataBlockSize {
		return nil, 0, fmt.Errorf("metadata block at %d has invalid size %d", pos, size)
	}
	data, err := sq.readAt(size, pos+2)
	if err != nil {
		return nil, 0, err
	}
	if h&mdUncompressed == 0 {
		if data, err = sq.decompress(data, metadataBlockSize); err != nil {
			return nil, 0, fmt.Errorf("metadata block at %d: %w", pos, err)
		}
	}
	b := metadataBlock{data: data, next: pos + 2 + size}
	sq.metadata[pos] = b
	return b.data, b.next, nil
}

// readLookupTable reads count entries of entrySize bytes from a table whose
// list of metadata block locations begins at start.
func (sq *Reader) readLookupTable(start, count uint64, entrySize int) ([]byte, error) {
	if count > sq.sb.BytesUsed {
		return nil, fmt.Errorf("invalid table entry count %d", count)
	}
	total := int(count) * entrySize
	blocks := (total + metadataBlockSize - 1) / metadataBlockSize
	index, err := sq.readAt(int64(blocks)*8, int64(start))
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, total)
	for i := 0; i < blocks; i++ {
		data, _, err := sq.readMetadataBlock(int64(binary.LittleEndian.Uint64(index[i*8:])))
		if err != nil {
			return nil, err
		}
		out = append(out, data...)
	}
	if len(out) < total {
		return nil, errors.New("table is truncated")
	}
	return out[:total], nil
}

// metadataReader reads a stream of bytes that may span metadata blocks.
type metadataReader struct {
	sq  *Reader
	pos int64
	buf []byte
}

// newMetadataReader returns a reader positioned at offset within the
// metadata block that begins block bytes after tableStart.
func (sq *Reader) newMetadataReader(tableStart int64, block uint64, offset uint16) (*metadataReader, error) {
	m := &metadataReader{sq: sq, pos: tableStart + int64(block)}
	if err := m.fill(); err != nil {
		return nil, err
	}
	if int(offset) > len(m.buf) {
		return nil, fmt.Errorf("metadata offset %d is outside its block", offset)
	}
	m.buf = m.buf[offset:]
	return m, nil
}

func (m *metadataReader) fill() error {
	data, next, err := m.sq.readMetadataBlock(m.pos)
	if err != nil {
		return err
	}
	m.buf, m.pos = data, next
	return nil
}

func (m *metadataReader) Read(p []byte) (int, error) {
	for len(m.buf) == 0 {
		if err := m.fill(); err != nil {
			return 0, err
		}
	}
	n := copy(p, m.buf)
	m.buf = m.buf[n:]
	return n, nil
}

type inode struct {
	typ    uint16
	mode   uint16
	uid    uint32
	gid    uint32
	mtime  uint32
	number uint32
	nlink  uint32
	xattr  uint32

	// directories
	dirBlock  uint32
	dirOffset uint16
	dirSize   uint32

	// regular files
	blocksStart uint64
	fileSize    uint64
	fragIdx     uint32
	fragOffset  uint32
	blockSizes  []uint32

	// symlinks
	target string

	// block and character devices
	rdev uint32
}

func (in *inode) isDir() bool {
	return in.typ == typeDir || in.typ == typeExtDir
}

// readInode reads the inode at ref, a metadata reference whose upper bits
// locate the metadata block within the inode table and whose lower 16 bits
// are the offset within that block.
func (sq *Reader) readInode(ref uint64) (*inode, error) {
	m, err := sq.newMetadataReader(int64(sq.sb.InodeTableStart), ref>>16, uint16(ref))
	if err != nil {
		return nil, fmt.Errorf("reading inode: %w", err)
	}
	var h struct {
		Type, Mode, UID, GID uint16
		Mtime, Number        uint32
	}
	if err := binary.Read(m, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("reading inode header: %w", err)
	}
	if int(h.UID) >= len(sq.ids) || int(h.GID) >= len(sq.ids) {
		return nil, fmt.Errorf("inode %d has an invalid uid or gid index", h.Number)
	}
	in := &inode{
		typ:     h.Type,
		mode:    h.Mode,
		uid:     sq.ids[h.UID],
		gid:     sq.ids[h.GID],
		mtime:   h.Mtime,
		number:  h.Number,
		nlink:   1,
		xattr:   noXattr,
		fragIdx: noFragment,
	}

	switch h.Type {
	case typeDir:
		var d struct {
			Block, Nlink   uint32
			Size, Offset   uint16
			ParentInodeNum uint32
		}
		err = binary.Read(m, binary.LittleEndian, &d)
		in.dirBlock, in.dirOffset, in.dirSize, in.nlink = d.Block, d.Offset, uint32(d.Size), d.Nlink
	case typeExtDir:
		var d struct {
			Nlink, Size, Block, ParentInodeNum uint32
			IndexCount, Offset                 uint16
			Xattr                              uint32
		}
		err = binary.Read(m, binary.LittleEndian, &d)
		in.dirBlock, in.dirOffset, in.dirSize, in.nlink, in.xattr = d.Block, d.Offset, d.Size, d.Nlink, d.Xattr
	case typeFile:
		var f struct {
			BlocksStart, FragIdx, FragOffset, Size uint32
		}
		if err = binary.Read(m, binary.LittleEndian, &f); err == nil {
			in.blocksStart, in.fileSize, in.fragIdx, in.fragOffset = uint64(f.BlocksStart), uint64(f.Size), f.FragIdx, f.FragOffset
			err = sq.readBlockSizes(m, in)
		}
	case typeExtFile:
		var f struct {
			BlocksStart, Size, Sparse         uint64
			Nlink, FragIdx, FragOffset, Xattr uint32
		}
		if err = binary.Read(m, binary.LittleEndian, &f); err == nil {
			in.blocksStart, in.fileSize, in.fragIdx, in.fragOffset = f.BlocksStart, f.Size, f.FragIdx, f.FragOffset
			in.nlink, in.xattr = f.Nlink, f.Xattr
			err = sq.readBlockSizes(m, in)
		}
	case typeSymlink, typeExtSymlink:
		var s struct {
			Nlink, Size uint32
		}
		if err = binary.Read(m, binary.LittleEndian, &s); err == nil {
			if s.Size == 0 || s.Size > 4096 {
				return nil, fmt.Errorf("inode %d has invalid symlink target size %d", h.Number, s.Size)
			}
			target := make([]byte, s.Size)
			if _, err = io.ReadFull(m, target); err == nil && h.Type == typeExtSymlink {
				err = binary.Read(m, binary.LittleEndian, &in.xattr)
			}
			in.target, in.nlink = string(target), s.Nlink
		}
	case typeBlockDev, typeCharDev:
		var d struct {
			Nlink, Rdev uint32
		}
		err = binary.Read(m, binary.LittleEndian, &d)
		in.nlink, in.rdev = d.Nlink, d.Rdev
	case typeExtBlockDev, typeExtCharDev:
		var d struct {
			Nlink, Rdev, Xattr uint32
		}
		err = binary.Read(m, binary.LittleEndian, &d)
		in.nlink, in.rdev, in.xattr = d.Nlink, d.Rdev, d.Xattr
	case typeFifo, typeSocket:
		err = binary.Read(m, binary.LittleEndian, &in.nlink)
	case typeExtFifo, typeExtSocket:
		var d struct {
			Nlink, Xattr uint32
		}
		err = binary.Read(m, binary.LittleEndian, &d)
		in.nlink, in.xattr = d.Nlink, d.Xattr
	default:
		return nil, fmt.Errorf("inode %d has unknown type %d", h.Number, h.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("reading inode %d: %w", h.Number, err)
	}
	return in, nil
}

func (sq *Reader) readBlockSizes(m *metadataReader, in *inode) error {
	bs := uint64(sq.sb.BlockSize)
	count := in.fileSize / bs
	if in.fragIdx == noFragment && in.fileSize%bs != 0 {
		count++
	}
	if count > maxBlockCount {
		return fmt.Errorf("file size %d is too large", in.fileSize)
	}
	in.blockSizes = make([]uint32, count)
	return binary.Read(m, binary.LittleEndian, in.blockSizes)
}

type dirEntry struct {
	name     string
	inodeRef uint64
}

// readDir returns the entries of a directory inode.
func (sq *Reader) readDir(in *inode) ([]dirEntry, error) {
	// The stored size includes three bytes for the implicit "." and ".."
	// entries.
	remaining := int64(in.dirSize) - 3
	if remaining <= 0 {
		return nil, nil
	}
	m, err := sq.newMetadataReader(int64(sq.sb.DirTableStart), uint64(in.dirBlock), in.dirOffset)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}
	var entries []dirEntry
	for remaining > 0 {
		var h struct {
			Count, Start, InodeNum uint32
		}
		if err := binary.Read(m, binary.LittleEndian, &h); err != nil {
			return nil, fmt.Errorf("reading directory header: %w", err)
		}
		remaining -= 12
		if h.Count >= 256 {
			return nil, fmt.Errorf("directory header has invalid entry count %d", h.Count+1)
		}
		for i := uint32(0); i <= h.Count; i++ {
			var e struct {
				Offset     uint16
				InodeDelta int16
				Type       uint16
				NameSize   uint16
			}
			if err := binary.Read(m, binary.LittleEndian, &e); err != nil {
				return nil, fmt.Errorf("reading directory entry: %w", err)
			}
			name := make([]byte, int(e.NameSize)+1)
			if _, err := io.ReadFull(m, name); err != nil {
				return nil, fmt.Errorf("reading directory entry: %w", err)
			}
			remaining -= 8 + int64(len(name))
			entries = append(entries, dirEntry{
				name:     string(name),
				inodeRef: uint64(h.Start)<<16 | uint64(e.Offset),
			})
		}
	}
	if remaining != 0 {
		return nil, errors.New("directory listing does not match its recorded size")
	}
	return entries, nil
}

// writeFile writes the contents of a regular file inode to w, which must be
// positioned at its start.
func (sq *Reader) writeFile(in *inode, w io.WriteSeeker) error {
	bs := int64(sq.sb.BlockSize)
	pos := int64(in.blocksStart)
	remaining := int64(in.fileSize)
	for _, s := range in.blockSizes {
		n := min(bs, remaining)
		size := int64(s &^ blockUncompressed)
		if size == 0 {
			// sparse block
			if _, err := w.Seek(n, io.SeekCurrent); err != nil {
				return err
			}
			remaining -= n
			continue
		}
		data, err := sq.readAt(size, pos)
		if err != nil {
			return err
		}
		pos += size
		if s&blockUncompressed == 0 {
			if data, err = sq.decompress(data, int(bs)); err != nil {
				return fmt.Errorf("data block at %d: %w", pos-size, err)
			}
		}
		if int64(len(data)) < n {
			return fmt.Errorf("data block at %d is truncated", pos-size)
		}
		if _, err := w.Write(data[:n]); err != nil {
			return err
		}
		remaining -= n
	}
	if remaining > 0 {
		if in.fragIdx == noFragment {
			return errors.New("file data is truncated")
		}
		data, err := sq.fragment(in.fragIdx)
		if err != nil {
			return err
		}
		off := int64(in.fragOffset)
		if off+remaining > int64(len(data)) {
			return errors.New("fragment is truncated")
		}
		if _, err := w.Write(data[off : off+remaining]); err != nil {
			return err
		}
	}
	return nil
}

// fragment returns the uncompressed contents of a fragment block.
func (sq *Reader) fragment(idx uint32) ([]byte, error) {
	if idx == sq.fragIdx {
		return sq.fragData, nil
	}
	if int(idx) >= len(sq.frags) {
		return nil, fmt.Errorf("invalid fragment index %d", idx)
	}
	f := sq.frags[idx]
	data, err := sq.readAt(int64(f.size&^blockUncompressed), int64(f.start))
	if err != nil {
		return nil, err
	}
	if f.size&blockUncompressed == 0 {
		if data, err = sq.decompress(data, int(sq.sb.BlockSize)); err != nil {
			return nil, fmt.Errorf("fragment %d: %w", idx, err)
		}
	}
	sq.fragIdx, sq.fragData = idx, data
	return data, nil
}

// xattrs returns the extended attributes at index idx of the xattr id table.
func (sq *Reader) xattrs(idx uint32) (map[string][]byte, error) {
	if idx == noXattr {
		return nil, nil
	}
	if int(idx) >= len(sq.xattrIDs) {
		return nil, fmt.Errorf("invalid xattr index %d", idx)
	}
	id := sq.xattrIDs[idx]
	m, err := sq.newMetadataReader(sq.xattrStart, id.ref>>16, uint16(id.ref))
	if err != nil {
		return nil, fmt.Errorf("reading xattrs: %w", err)
	}
	out := make(map[string][]byte, id.count)
	for i := uint32(0); i < id.count; i++ {
		var k struct {
			Type, NameSize uint16
		}
		if err := binary.Read(m, binary.LittleEndian, &k); err != nil {
			return nil, fmt.Errorf("reading xattr key: %w", err)
		}
		name := make([]byte, k.NameSize)
		if _, err := io.ReadFull(m, name); err != nil {
			return nil, fmt.Errorf("reading xattr key: %w", err)
		}
		var prefix string
		switch k.Type & 0xff {
		case 0:
			prefix = "user."
		case 1:
			prefix = "trusted."
		case 2:
			prefix = "security."
		default:
			return nil, fmt.Errorf("unknown xattr prefix type %d", k.Type&0xff)
		}
		value, err := readXattrValue(m)
		if err != nil {
			return nil, err
		}
		if k.Type&0x100 != 0 {
			// The value is stored out of line; this value is a reference
			// to it.
			if len(value) != 8 {
				return nil, errors.New("invalid out-of-line xattr reference")
			}
			ref := binary.LittleEndian.Uint64(value)
			vm, err := sq.newMetadataReader(sq.xattrStart, ref>>16, uint16(ref))
			if err != nil {
				return nil, fmt.Errorf("reading xattr value: %w", err)
			}
			if value, err = readXattrValue(vm); err != nil {
				return nil, err
			}
		}
		out[prefix+string(name)] = value
	}
	return out, nil
}

func readXattrValue(r io.Reader) ([]byte, error) {
	var size uint32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, fmt.Errorf("reading xattr value: %w", err)
	}
	if size > maxXattrValueSize {
		return nil, fmt.Errorf("xattr value size %d is too large", size)
	}
	value := make([]byte, size)
	if _, err := io.ReadFull(r, value); err != nil {
		return nil, fmt.Errorf("reading xattr value: %w", err)
	}
	return value, nil
}
