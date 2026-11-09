package overlay

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// writeELF writes a minimal 64-bit little-endian ELF file containing only a
// valid ELF header with the given machine type. debug/elf only needs the
// header to report the machine type.
func writeELF(t *testing.T, path string, machine elf.Machine) {
	t.Helper()
	var buf bytes.Buffer
	// e_ident: magic, ELFCLASS64, little-endian, current version, SysV, padding
	buf.Write([]byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	hdr := struct {
		Type      uint16
		Machine   uint16
		Version   uint32
		Entry     uint64
		Phoff     uint64
		Shoff     uint64
		Flags     uint32
		Ehsize    uint16
		Phentsize uint16
		Phnum     uint16
		Shentsize uint16
		Shnum     uint16
		Shstrndx  uint16
	}{
		Machine:   uint16(machine),
		Version:   1,
		Ehsize:    64,
		Phentsize: 56,
		Shentsize: 64,
	}
	assert.NoError(t, binary.Write(&buf, binary.LittleEndian, hdr))
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, buf.Bytes(), 0o755))
}

func Test_elfMachine(t *testing.T) {
	dir := t.TempDir()

	x86 := filepath.Join(dir, "wwclient.x86_64")
	writeELF(t, x86, elf.EM_X86_64)
	machine, err := elfMachine(x86)
	assert.NoError(t, err)
	assert.Equal(t, elf.EM_X86_64, machine)

	aarch64 := filepath.Join(dir, "wwclient.aarch64")
	writeELF(t, aarch64, elf.EM_AARCH64)
	machine, err = elfMachine(aarch64)
	assert.NoError(t, err)
	assert.Equal(t, elf.EM_AARCH64, machine)

	// A non-ELF file (like a text file or a Mach-O binary) reports an error
	// rather than a bogus machine type.
	text := filepath.Join(dir, "notelf")
	assert.NoError(t, os.WriteFile(text, []byte("#!/bin/sh\n"), 0o755))
	_, err = elfMachine(text)
	assert.Error(t, err)

	_, err = elfMachine(filepath.Join(dir, "missing"))
	assert.Error(t, err)
}

func Test_resolveWithinRoot(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// A relative symlink chain, as found in glibc images:
	// /bin -> usr/bin and /usr/bin/sh -> bash
	env.WriteFile("image/usr/bin/bash", "#!/bin/sh\n")
	env.WriteFile("image/usr/bin/env", "#!/bin/sh\n")
	env.Symlink("usr/bin", "image/bin")
	env.Symlink("bash", "image/usr/bin/sh")

	// An absolute symlink inside the image must resolve within the image
	// root, not against the server's own /.
	env.Symlink("/usr/bin/bash", "image/usr/bin/absolute-sh")

	resolved, err := resolveWithinRoot(env.GetPath("image"), "usr/bin/env")
	assert.NoError(t, err)
	assert.Equal(t, env.GetPath("image/usr/bin/env"), resolved)

	resolved, err = resolveWithinRoot(env.GetPath("image"), "bin/sh")
	assert.NoError(t, err)
	assert.Equal(t, env.GetPath("image/usr/bin/bash"), resolved)

	resolved, err = resolveWithinRoot(env.GetPath("image"), "usr/bin/absolute-sh")
	assert.NoError(t, err)
	assert.Equal(t, env.GetPath("image/usr/bin/bash"), resolved)

	// Missing files report an error.
	_, err = resolveWithinRoot(env.GetPath("image"), "usr/bin/missing")
	assert.Error(t, err)

	// Symlink loops report an error instead of hanging.
	env.Symlink("loop", "image/usr/bin/loop")
	_, err = resolveWithinRoot(env.GetPath("image"), "usr/bin/loop")
	assert.Error(t, err)
}

// Test_resolveWithinRoot_chroot checks that resolution stays inside the
// image root even when symlinks would make the kernel follow HOST paths.
func Test_resolveWithinRoot_chroot(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// An absolute symlink to a directory, followed by an absolute symlink
	// to a file through that directory: exactly the chain that makes a
	// naive path.Join(root, target) read the server's own files.
	writeELF(t, env.GetPath("image/usr/bin/bash"), elf.EM_AARCH64)
	env.Symlink("/usr/bin", "image/bin")
	env.Symlink("/bin/bash", "image/usr/bin/sh")

	resolved, err := resolveWithinRoot(env.GetPath("image"), "usr/bin/sh")
	assert.NoError(t, err)
	assert.Equal(t, env.GetPath("image/usr/bin/bash"), resolved)

	// The same chain starting at the symlinked directory.
	resolved, err = resolveWithinRoot(env.GetPath("image"), "bin/sh")
	assert.NoError(t, err)
	assert.Equal(t, env.GetPath("image/usr/bin/bash"), resolved)

	// A symlink to an absolute path that exists on any Linux/macOS host
	// but NOT in the image must fail, never return a host path.
	env.Symlink("/etc/hosts", "image/usr/bin/hosts")
	_, err = resolveWithinRoot(env.GetPath("image"), "usr/bin/hosts")
	assert.Error(t, err)

	// A two-way symlink loop reports an error instead of hanging.
	env.Symlink("b", "image/a")
	env.Symlink("a", "image/b")
	_, err = resolveWithinRoot(env.GetPath("image"), "a")
	assert.Error(t, err)
}

// Test_resolveWithinRoot_dotdot checks chroot `..` handling in a layout
// without symlinked directories.
func Test_resolveWithinRoot_dotdot(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// `..` is clamped at the image root (chroot semantics).
	writeELF(t, env.GetPath("image/bin/env"), elf.EM_AARCH64)
	env.Symlink("../../../../bin/env", "image/usr/bin/env")
	resolved, err := resolveWithinRoot(env.GetPath("image"), "usr/bin/env")
	assert.NoError(t, err)
	assert.Equal(t, env.GetPath("image/bin/env"), resolved)

	// A `..`-only symlink must not escape the image root.
	env.Symlink("../../..", "image/usr/bin/up")
	_, err = resolveWithinRoot(env.GetPath("image"), "usr/bin/up")
	assert.Error(t, err)
}

func Test_imageELFMachine(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// No image: error, no panic.
	_, err := imageELFMachine("rocky9")
	assert.Error(t, err)

	writeELF(t, env.GetPath("var/lib/warewulf/chroots/rocky9/rootfs/usr/bin/env"), elf.EM_AARCH64)
	machine, err := imageELFMachine("rocky9")
	assert.NoError(t, err)
	assert.Equal(t, elf.EM_AARCH64, machine)

	// Falls back to bin/sh through the /bin -> usr/bin symlink.
	writeELF(t, env.GetPath("var/lib/warewulf/chroots/centos7/rootfs/usr/bin/sh"), elf.EM_X86_64)
	env.Symlink("usr/bin", "var/lib/warewulf/chroots/centos7/rootfs/bin")
	machine, err = imageELFMachine("centos7")
	assert.NoError(t, err)
	assert.Equal(t, elf.EM_X86_64, machine)
}

func Test_machineName(t *testing.T) {
	assert.Equal(t, "x86_64", machineName(elf.EM_X86_64))
	assert.Equal(t, "aarch64", machineName(elf.EM_AARCH64))
	assert.NotEmpty(t, machineName(elf.Machine(255)))
}

// Test_warnWWClientArch checks the warning fires on an architecture
// mismatch, and stays quiet on a match or when the image can't be probed.
func Test_warnWWClientArch(t *testing.T) {
	cases := map[string]struct {
		overlayMachine elf.Machine
		imageMachine   elf.Machine
		withImage      bool
		expectWarning  bool
	}{
		"mismatch warns": {
			overlayMachine: elf.EM_X86_64,
			imageMachine:   elf.EM_AARCH64,
			withImage:      true,
			expectWarning:  true,
		},
		"match is quiet": {
			overlayMachine: elf.EM_AARCH64,
			imageMachine:   elf.EM_AARCH64,
			withImage:      true,
			expectWarning:  false,
		},
		"missing image is quiet": {
			overlayMachine: elf.EM_X86_64,
			withImage:      false,
			expectWarning:  false,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()

			writeELF(t, env.GetPath("var/lib/warewulf/overlays/wwclient/rootfs/warewulf/wwclient"), tt.overlayMachine)
			if tt.withImage {
				writeELF(t, env.GetPath("var/lib/warewulf/chroots/rocky9/rootfs/usr/bin/env"), tt.imageMachine)
			}

			buf := new(bytes.Buffer)
			wwlog.SetLogWriter(buf)
			defer wwlog.SetLogWriter(os.Stderr)

			nodeConf := node.NewNode("node1")
			nodeConf.ImageName = "rocky9"
			assert.NotPanics(t, func() { warnWWClientArch(nodeConf, env.GetPath("var/lib/warewulf/overlays/wwclient/rootfs")) })

			output := buf.String()
			if tt.expectWarning {
				assert.Contains(t, output, "node1")
				assert.Contains(t, output, "x86_64")
				assert.Contains(t, output, "aarch64")
				assert.Contains(t, output, "wwclient.aarch64")
			} else {
				assert.NotContains(t, output, "wwclient.aarch64")
				assert.NotContains(t, output, "WARN")
			}
		})
	}
}

// Test_warnWWClientArch_uncommonArch checks the advice for image
// architectures without a prebuilt wwclient overlay.
func Test_warnWWClientArch_uncommonArch(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	writeELF(t, env.GetPath("var/lib/warewulf/overlays/wwclient/rootfs/warewulf/wwclient"), elf.EM_X86_64)
	writeELF(t, env.GetPath("var/lib/warewulf/chroots/riscv64/rootfs/usr/bin/env"), elf.EM_RISCV)

	buf := new(bytes.Buffer)
	wwlog.SetLogWriter(buf)
	defer wwlog.SetLogWriter(os.Stderr)

	nodeConf := node.NewNode("node1")
	nodeConf.ImageName = "riscv64"
	warnWWClientArch(nodeConf, env.GetPath("var/lib/warewulf/overlays/wwclient/rootfs"))

	output := buf.String()
	assert.Contains(t, output, "no prebuilt wwclient overlay exists for riscv64")
	assert.NotContains(t, output, "wwclient.riscv64")
}

// Test_BuildOverlay_WWClientArchWarning verifies the check is wired into
// BuildOverlay for node builds.
func Test_BuildOverlay_WWClientArchWarning(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	writeELF(t, env.GetPath("var/lib/warewulf/overlays/wwclient/rootfs/warewulf/wwclient"), elf.EM_X86_64)
	writeELF(t, env.GetPath("var/lib/warewulf/chroots/rocky9/rootfs/usr/bin/env"), elf.EM_AARCH64)

	buf := new(bytes.Buffer)
	wwlog.SetLogWriter(buf)
	defer wwlog.SetLogWriter(os.Stderr)

	nodeConf := node.NewNode("node1")
	nodeConf.ImageName = "rocky9"
	err := BuildOverlay(nodeConf, []node.Node{nodeConf}, "system", []string{"wwclient"})
	assert.NoError(t, err)
	assert.True(t, strings.Contains(buf.String(), "wwclient.aarch64"),
		"expected an architecture mismatch warning, got: %s", buf.String())
}
