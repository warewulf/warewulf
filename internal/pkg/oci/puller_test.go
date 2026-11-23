package oci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestGetReference(t *testing.T) {
	temp, err := os.MkdirTemp(os.TempDir(), "ww-archive-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(temp) }()

	tests := []struct {
		name string
		uri  string
		err  error
	}{
		{
			name: "file archive ok case",
			uri:  "test.tar",
			err:  nil,
		},
		{
			name: "file archive ko case, because having colon in file name",
			uri:  "test:latest.tar",
			err:  fmt.Errorf("/test:latest.tar should not contain a colon"),
		},
		{
			name: "file archive ko case, because having colon in path name",
			uri:  "/test:/test.tar",
			err:  fmt.Errorf("/test:/test.tar should not contain a colon"),
		},
		{
			name: "docker ok case",
			uri:  "docker://test:latest",
			err:  nil,
		},
		{
			name: "docker daemon ok case",
			uri:  "docker-daemon://test:latest",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Logf("running test: %s", tt.name)
		if !strings.HasPrefix(tt.uri, "docker") && !strings.HasPrefix(tt.uri, "docker-daemon") {
			tt.uri = filepath.Join(temp, tt.uri)
			parent := filepath.Dir(tt.uri)
			err := os.MkdirAll(parent, 0o755)
			assert.NoError(t, err)
			f, err := os.Create(tt.uri)
			assert.NoError(t, err)
			assert.NoError(t, f.Close())
		}

		_, err := getReference(tt.uri)
		if tt.err == nil {
			assert.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, tt.err.Error())
		}
	}

}

func TestUnpackError(t *testing.T) {
	// the shape of the error umoci returns when Lremovexattr fails
	xattrErr := fmt.Errorf("unpack layer: unpack entry: .profile: apply hdr metadata: clear xattr metadata: /chroots/rocky/rootfs/.profile: %w",
		fmt.Errorf("lclearxattrs: remove xattr %q: %w", "system.nfs4_dacl", unix.EINVAL))

	tests := []struct {
		name    string
		err     error
		hint    bool
		wrapped error
	}{
		{
			name:    "xattr removal refused",
			err:     xattrErr,
			hint:    true,
			wrapped: unix.EINVAL,
		},
		{
			name: "other EINVAL",
			err:  fmt.Errorf("unpack layer: %w", unix.EINVAL),
			hint: false,
		},
		{
			name: "other xattr error",
			err:  fmt.Errorf("lclearxattrs: get list: %w", unix.EIO),
			hint: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := unpackError("/chroots/rocky/rootfs", tt.err)
			assert.ErrorContains(t, err, "unable to unpack rootfs: "+tt.err.Error())
			if tt.hint {
				assert.ErrorContains(t, err, "the filesystem holding /chroots/rocky/rootfs refused to remove an extended attribute")
				assert.ErrorContains(t, err, "paths:wwchrootdir")
				assert.ErrorIs(t, err, tt.wrapped)
			} else {
				assert.NotContains(t, err.Error(), "paths:wwchrootdir")
			}
		})
	}
}
