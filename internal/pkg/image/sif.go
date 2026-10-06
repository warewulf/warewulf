package image

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/apptainer/sif/v2/pkg/sif"
	securejoin "github.com/cyphar/filepath-securejoin"

	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// The SIF global header begins with a 32-byte launch script followed by
// this magic value.
var sifMagic = []byte("SIF_MAGIC\x00")

const sifMagicOffset = 32

// squashfsMagic begins a squashfs superblock.
var squashfsMagic = []byte("hsqs")

// ErrUnsupportedSIF is returned when the contents of a SIF image cannot be
// imported.
var ErrUnsupportedSIF = errors.New("unsupported SIF image")

// IsSIF reports whether the file at path is a SIF image.
func IsSIF(path string) (bool, error) {
	// Only regular files are opened: opening a FIFO or a device can block
	// or have side effects.
	if fi, err := os.Stat(path); err != nil || !fi.Mode().IsRegular() {
		return false, err
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}

	defer func() {
		_ = f.Close()
	}()

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
// as the image name. Only squashfs partitions are supported. The partition
// is extracted with unsquashfs.
func ImportSIF(uri string, name string) error {
	if !ValidName(name) || name == "." || name == ".." {
		return errors.New("Image name contains illegal characters: " + name)
	}

	unsquashfs, err := exec.LookPath("unsquashfs")
	if err != nil {
		return fmt.Errorf("importing SIF images requires unsquashfs (squashfs-tools): %w", err)
	}

	f, err := os.Open(uri)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	// The container is not unloaded: that would only close f, which the
	// deferred Close above does.
	img, err := sif.LoadContainer(f)
	if err != nil {
		return fmt.Errorf("%w: could not load SIF image %s: %w", ErrUnsupportedSIF, uri, err)
	}

	d, err := img.GetDescriptor(sif.WithPartitionType(sif.PartPrimSys))
	if err != nil {
		if _, ociErr := img.GetDescriptor(sif.WithDataType(sif.DataOCIRootIndex)); ociErr == nil {
			return fmt.Errorf("%w: %s is an OCI-SIF image, which is not supported", ErrUnsupportedSIF, uri)
		}
		return fmt.Errorf("%w: could not find the primary system partition in %s: %w", ErrUnsupportedSIF, uri, err)
	}
	fsType, _, arch, err := d.PartitionMetadata()
	if err != nil {
		return fmt.Errorf("%w: could not read partition metadata in %s: %w", ErrUnsupportedSIF, uri, err)
	}
	if fsType != sif.FsSquash {
		return fmt.Errorf("%w: %s has an unsupported primary partition filesystem: %v", ErrUnsupportedSIF, uri, fsType)
	}
	magic := make([]byte, len(squashfsMagic))
	if _, err := f.ReadAt(magic, d.Offset()); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("could not read %s: %w", uri, err)
	} else if err != nil || !bytes.Equal(magic, squashfsMagic) {
		return fmt.Errorf("%w: primary partition in %s is not a squashfs filesystem", ErrUnsupportedSIF, uri)
	}
	if overlays, _ := img.GetDescriptors(sif.WithPartitionType(sif.PartOverlay)); len(overlays) > 0 {
		wwlog.Warn("%s has an overlay partition, which is not imported", uri)
	}
	wwlog.Info("Importing SIF primary partition: arch=%s offset=%d size=%d", arch, d.Offset(), d.Size())

	fullPath := RootFsDir(name)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return err
	}

	args := []string{"-f", "-n", "-d", fullPath}
	if os.Geteuid() != 0 {
		// unsquashfs fails on xattrs that only root can set.
		args = append(args, "-no-xattrs")
	}
	hasOffset := unsquashfsHasOffset(unsquashfs)
	if hasOffset {
		args = append(args, "-o", strconv.FormatInt(d.Offset(), 10), uri)
	} else {
		// unsquashfs before 4.4 (e.g., EL8) has no -o, so the partition is
		// copied to a file of its own.
		tmp, err := os.CreateTemp(filepath.Dir(fullPath), "sif-*.squashfs")
		if err != nil {
			return err
		}
		defer func() {
			_ = os.Remove(tmp.Name())
		}()
		_, err = io.Copy(tmp, d.GetReader())
		if err := errors.Join(err, tmp.Close()); err != nil {
			return err
		}
		args = append(args, tmp.Name())
	}
	wwlog.Debug("%s %s", unsquashfs, strings.Join(args, " "))
	out, err := exec.Command(unsquashfs, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("unsquashfs failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	wwlog.Debug("unsquashfs: %s", strings.TrimSpace(string(out)))
	// unsquashfs before 4.4 exits 0 after non-fatal errors, which are only
	// reported in its output.
	if !hasOffset && (bytes.Contains(out, []byte("fail")) || bytes.Contains(out, []byte("could not"))) {
		wwlog.Warn("unsquashfs reported errors importing %s: %s", uri, strings.TrimSpace(string(out)))
	}

	// Resolve symlinks such as /bin -> /usr/bin inside the image, not on the
	// host.
	sh, err := securejoin.SecureJoin(fullPath, "bin/sh")
	if fi, statErr := os.Stat(sh); err != nil || statErr != nil || !fi.Mode().IsRegular() {
		return fmt.Errorf("%w: %s has no /bin/sh", ErrUnsupportedSIF, uri)
	}

	// As root, unsquashfs restores the SELinux labels of the system that
	// built the image, so they are reset to the local policy.
	if os.Geteuid() == 0 {
		return util.RestoreSELinuxContext(fullPath)
	}
	return nil
}

var unsquashfsVersion = regexp.MustCompile(`version (\d+)\.(\d+)`)

// unsquashfsHasOffset reports whether unsquashfs supports -o, which was
// added in squashfs-tools 4.4.
func unsquashfsHasOffset(unsquashfs string) bool {
	out, _ := exec.Command(unsquashfs, "-version").CombinedOutput()
	m := unsquashfsVersion.FindSubmatch(out)
	if m == nil {
		return false
	}
	major, _ := strconv.Atoi(string(m[1]))
	minor, _ := strconv.Atoi(string(m[2]))
	return major > 4 || (major == 4 && minor >= 4)
}
