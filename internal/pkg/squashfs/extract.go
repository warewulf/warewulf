package squashfs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type extractor struct {
	sq   *Reader
	root bool
	// links maps inode numbers of files with more than one link to the
	// path they were first extracted to.
	links map[uint32]string
}

// Extract writes the contents of the filesystem into dst, which must be an
// existing directory. Entries that already exist in dst are replaced.
//
// Ownership, device nodes, and non-user extended attributes are only
// restored when running as root.
func (sq *Reader) Extract(dst string) error {
	root, err := sq.readInode(sq.sb.RootInode)
	if err != nil {
		return err
	}
	if !root.isDir() {
		return errors.New("squashfs root inode is not a directory")
	}
	x := &extractor{
		sq:    sq,
		root:  os.Geteuid() == 0,
		links: map[uint32]string{},
	}
	return x.extractDir(root, dst)
}

func (x *extractor) extractDir(in *inode, dir string) error {
	entries, err := x.sq.readDir(in)
	if err != nil {
		return fmt.Errorf("%s: %w", dir, err)
	}
	for _, e := range entries {
		if e.name == "." || e.name == ".." || strings.ContainsAny(e.name, "/\x00") {
			return fmt.Errorf("%s: invalid entry name %q", dir, e.name)
		}
		child, err := x.sq.readInode(e.inodeRef)
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.Join(dir, e.name), err)
		}
		if err := x.extract(child, filepath.Join(dir, e.name)); err != nil {
			return err
		}
	}
	// Attributes are set after the children are written so that read-only
	// directories can be populated and modification times are preserved.
	return x.setAttrs(in, dir)
}

func (x *extractor) extract(in *inode, path string) error {
	exists, err := prepare(path, in.isDir())
	if err != nil {
		return err
	}

	if in.isDir() {
		if !exists {
			if err := os.Mkdir(path, 0o700); err != nil {
				return err
			}
		}
		return x.extractDir(in, path)
	}

	if in.nlink > 1 {
		if first, ok := x.links[in.number]; ok {
			return os.Link(first, path)
		}
		x.links[in.number] = path
	}

	switch in.typ {
	case typeFile, typeExtFile:
		if err := x.writeFile(in, path); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	case typeSymlink, typeExtSymlink:
		if err := os.Symlink(in.target, path); err != nil {
			return err
		}
	case typeBlockDev, typeExtBlockDev, typeCharDev, typeExtCharDev:
		if !x.root {
			wwlog.Warn("Skipping device node %s: creating device nodes requires root", path)
			delete(x.links, in.number)
			return nil
		}
		mode := uint32(unix.S_IFCHR)
		if in.typ == typeBlockDev || in.typ == typeExtBlockDev {
			mode = unix.S_IFBLK
		}
		major := (in.rdev >> 8) & 0xfff
		minor := (in.rdev & 0xff) | ((in.rdev >> 12) & 0xfff00)
		if err := unix.Mknod(path, mode|0o600, int(unix.Mkdev(major, minor))); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	case typeFifo, typeExtFifo:
		if err := unix.Mkfifo(path, 0o600); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	case typeSocket, typeExtSocket:
		if err := unix.Mknod(path, unix.S_IFSOCK|0o600, 0); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return x.setAttrs(in, path)
}

func (x *extractor) writeFile(in *inode, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if err := x.sq.writeFile(in, f); err != nil {
		_ = f.Close()
		return err
	}
	// Extend the file over any trailing sparse blocks.
	if err := f.Truncate(int64(in.fileSize)); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// prepare removes any existing entry at path unless both it and the entry
// being extracted are directories. It reports whether a directory was kept.
func prepare(path string, isDir bool) (bool, error) {
	fi, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if isDir && fi.IsDir() {
		return true, nil
	}
	return false, os.RemoveAll(path)
}

// setAttrs applies ownership, extended attributes, permissions, and
// modification time. Ownership is set first because chown clears setuid and
// setgid bits and file capabilities.
func (x *extractor) setAttrs(in *inode, path string) error {
	if x.root {
		if err := os.Lchown(path, int(in.uid), int(in.gid)); err != nil {
			return err
		}
	}

	xattrs, err := x.sq.xattrs(in.xattr)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	for name, value := range xattrs {
		if !x.root && !strings.HasPrefix(name, "user.") {
			wwlog.Debug("Skipping xattr %s on %s: requires root", name, path)
			continue
		}
		if err := unix.Lsetxattr(path, name, value, 0); err != nil {
			return fmt.Errorf("%s: setting xattr %s: %w", path, name, err)
		}
	}

	if in.typ != typeSymlink && in.typ != typeExtSymlink {
		if err := unix.Chmod(path, uint32(in.mode&0o7777)); err != nil {
			return err
		}
	}

	ts := unix.NsecToTimespec(int64(in.mtime) * 1e9)
	if err := unix.UtimesNanoAt(unix.AT_FDCWD, path, []unix.Timespec{ts, ts}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
