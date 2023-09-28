package kernel

import (
	"os"
	"path"
	"testing"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/stretchr/testify/assert"
)

var kernelBuildTests = []struct {
	kernelVersion  string
	kernelName     string
	kernelFileName string
	succeed        bool
}{
	{"4.3.2.1", "kernel1", "vmlinuz-1.2.3.4.gz", false},
	{"1.2.3.4", "kernel1", "vmlinuz-1.2.3.4.gz", true},
}

func Test_BuildKernel(t *testing.T) {
	wwlog.SetLogLevel(wwlog.DEBUG)
	srvDir, err := os.MkdirTemp(os.TempDir(), "ww-test-srv-*")
	assert.NoError(t, err)
	defer os.RemoveAll(srvDir)
	conf := warewulfconf.Get()
	conf.Paths.WWProvisiondir = srvDir
	kernelDir, err := os.MkdirTemp(os.TempDir(), "ww-test-kernelImage-*")
	assert.NoError(t, err)
	defer os.RemoveAll(kernelDir)
	{
		err = os.MkdirAll(path.Join(kernelDir, "boot"), 0755)
		assert.NoError(t, err)
		for _, tt := range kernelBuildTests {
			_, err = os.Create(path.Join(kernelDir, "boot", tt.kernelFileName))
			assert.NoError(t, err)
			err = os.MkdirAll(path.Join(kernelDir, "lib/modules", tt.kernelVersion), 0755)
			assert.NoError(t, err)
			_, err = os.Create(path.Join(kernelDir, "lib/modules", tt.kernelVersion, "test-module"))
			assert.NoError(t, err)
			err = os.MkdirAll(path.Join(kernelDir, "lib/firmware"), 0755)
			assert.NoError(t, err)
			_, err = os.Create(path.Join(kernelDir, "lib/firmware/test-firmware"))
			assert.NoError(t, err)
		}
	}
	for _, tt := range kernelBuildTests {
		t.Run(tt.kernelName, func(t *testing.T) {
			err = Build(tt.kernelVersion, tt.kernelName, kernelDir)
			if tt.succeed {
				assert.NoError(t, err)
				assert.FileExists(t, path.Join(srvDir, "kernel", tt.kernelName, "vmlinuz"))
				assert.FileExists(t, path.Join(srvDir, "kernel", tt.kernelName, "kmods.img.gz"))
				assert.FileExists(t, path.Join(srvDir, "kernel", tt.kernelName, "kmods.img"))
				files, err := util.CpioFiles(path.Join(srvDir, "kernel", tt.kernelName, "kmods.img"))
				assert.NoError(t, err)
				assert.Equal(t, files, []string{"lib/firmware/test-firmware", "lib/modules/1.2.3.4/test-module"})
			} else {
				assert.Error(t, err)
			}
		})
	}
}
