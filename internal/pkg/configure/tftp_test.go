package configure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyGrub(t *testing.T) {
	for arch, shimLoader := range map[string]string{
		"amd64": "grubx64.efi",
		"arm64": "grubaa64.efi",
	} {
		t.Run(arch, func(t *testing.T) {
			src := filepath.Join(t.TempDir(), "grub.efi")
			require.NoError(t, os.WriteFile(src, []byte(arch), 0o644))
			dst := t.TempDir()
			require.NoError(t, copyGrub(src, dst, arch))

			files, err := os.ReadDir(dst)
			require.NoError(t, err)
			assert.Len(t, files, 2)
			for _, name := range []string{"grub.efi", shimLoader} {
				got, err := os.ReadFile(filepath.Join(dst, name))
				require.NoError(t, err)
				assert.Equal(t, arch, string(got))
			}
		})
	}
}
