package overlay

import (
	"debug/elf"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
When an overlay build contains a wwclient binary, its architecture should
match the architecture of the node's image. A mismatched binary fails on
the node with "Exec format error" and nothing on the server reports why.

warnWWClientArch compares the ELF machine type of the wwclient binary in a
built overlay directory against that of a well-known binary in the node's
image, and warns when they differ. It never fails the build: if anything
cannot be read, or is not an ELF file, it stays quiet (debug log at most).
*/
func warnWWClientArch(nodeConf node.Node, buildDir string) {
	nodeId := nodeConf.Id()
	imageName := nodeConf.ImageName
	if imageName == "" {
		return
	}

	conf := warewulfconf.Get()
	overlayClient := path.Join(buildDir, conf.Paths.WWClientdir, "wwclient")
	clientMachine, err := elfMachine(overlayClient)
	if err != nil {
		wwlog.Debug("skipping wwclient architecture check: cannot read %s: %s", overlayClient, err)
		return
	}

	imageMachine, err := imageELFMachine(imageName)
	if err != nil {
		wwlog.Debug("skipping wwclient architecture check for node %s: cannot probe image %s: %s", nodeId, imageName, err)
		return
	}

	if clientMachine != imageMachine {
		clientArch := machineName(clientMachine)
		imageArch := machineName(imageMachine)
		var advice string
		switch imageArch {
		case "x86_64", "aarch64":
			advice = fmt.Sprintf("assign the wwclient.%s overlay instead of wwclient", imageArch)
		default:
			advice = fmt.Sprintf("no prebuilt wwclient overlay exists for %s, so a wwclient built for that architecture must be provided in its overlay", imageArch)
		}
		wwlog.Warn("node %s: the wwclient binary in its overlay is %s, but image %s is %s: %s, or the node will fail to run wwclient",
			nodeId, clientArch, imageName, imageArch, advice)
	}
}

/*
imageELFMachine returns the ELF machine type of the image's own
architecture, read from a well-known binary inside the image rootfs.
*/
func imageELFMachine(imageName string) (elf.Machine, error) {
	rootfs := image.RootFsDir(imageName)
	for _, candidate := range []string{"usr/bin/env", "usr/bin/sh", "bin/sh"} {
		resolved, err := resolveWithinRoot(rootfs, candidate)
		if err != nil {
			continue
		}
		machine, err := elfMachine(resolved)
		if err != nil {
			continue
		}
		return machine, nil
	}
	return 0, os.ErrNotExist
}

/*
elfMachine returns the ELF machine type of the file at path.
*/
func elfMachine(path string) (elf.Machine, error) {
	f, err := elf.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	return f.Machine, nil
}

/*
resolveWithinRoot resolves relPath (and any symlinks along it) relative to
root, never relative to the server's own file system. It walks one component
at a time over a work queue of remaining components, keeping a resolved
prefix that contains no symlinks, so no path with unresolved intermediate
components is ever handed to the kernel: an intermediate symlink in the image
(e.g. /bin -> /usr/bin) can never make the kernel follow a HOST path. `..` is
clamped at the image root (chroot semantics). The result is an absolute path
within root whose components are all real, non-symlink entries.
*/
func resolveWithinRoot(root string, relPath string) (string, error) {
	const maxSymlinkHops = 40

	queue := strings.Split(strings.TrimPrefix(path.Clean("/"+relPath), "/"), "/")
	var prefix []string // resolved, symlink-free components below root
	hops := 0
	for len(queue) > 0 {
		part := queue[0]
		queue = queue[1:]
		switch part {
		case "", ".":
			continue
		case "..":
			// Chroot semantics: `..` at the image root stays at the root.
			if len(prefix) > 0 {
				prefix = prefix[:len(prefix)-1]
			}
			continue
		}
		full := filepath.Join(root, filepath.Join(prefix...), part)
		info, err := os.Lstat(full)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			prefix = append(prefix, part)
			continue
		}
		hops++
		if hops > maxSymlinkHops {
			// Too many levels of symlink resolution: treat as a loop.
			return "", errors.New("too many levels of symbolic links: " + relPath)
		}
		target, err := os.Readlink(full)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(target, "/") {
			// Absolute symlink targets point at the root of the image.
			prefix = prefix[:0]
			target = strings.TrimPrefix(path.Clean(target), "/")
		}
		// Resolve the target's components before anything else. For a
		// relative target, `..` must stay relative to the symlink's own
		// directory (the current prefix), so clean without a leading "/".
		targetParts := strings.Split(path.Clean(target), "/")
		queue = append(targetParts, queue...)
	}
	if len(prefix) == 0 {
		return "", os.ErrInvalid
	}
	return filepath.Join(root, filepath.Join(prefix...)), nil
}

/*
machineName renders an ELF machine type in the terms used by uname and the
per-architecture overlay names.
*/
func machineName(m elf.Machine) string {
	switch m {
	case elf.EM_X86_64:
		return "x86_64"
	case elf.EM_AARCH64:
		return "aarch64"
	case elf.EM_386:
		return "i686"
	case elf.EM_ARM:
		return "armv7hl"
	case elf.EM_PPC64:
		return "ppc64le"
	case elf.EM_RISCV:
		return "riscv64"
	case elf.EM_S390:
		return "s390x"
	default:
		return m.String()
	}
}
