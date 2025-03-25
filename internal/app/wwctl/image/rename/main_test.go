package rename

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	imageList "github.com/warewulf/warewulf/internal/app/wwctl/image/list"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Rename(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile(path.Join(testenv.WWChrootdir, "test-image/rootfs/file"), `test`)
	env.WriteFile("etc/warewulf/nodes.conf", `
nodeprofiles:
  default:
    image name: test-image
nodes:
  n1:
    image name: test-image`)
	defer env.RemoveAll()
	warewulfd.SetNoDaemon()

	// first we will verify that there is an existing image
	t.Run("image list", func(t *testing.T) {
		verifyImageListOutput(t, "test-image")
	})

	// then rename it
	t.Run("image rename", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetOut(os.Stdout)
		baseCmd.SetErr(os.Stdout)
		baseCmd.SetArgs([]string{"test-image", "test-image-rename"})
		err := baseCmd.Execute()
		assert.NoError(t, err)
	})

	// retrieve again
	t.Run("Image list", func(t *testing.T) {
		verifyImageListOutput(t, "test-image-rename")
	})

	assert.YAMLEq(t, `
nodeprofiles:
  default:
    image name: test-image-rename
nodes:
  n1:
    image name: test-image-rename`, env.ReadFile("etc/warewulf/nodes.conf"))
}

func verifyImageListOutput(t *testing.T, content string) {
	baseCmd := imageList.GetCommand()
	buf := new(bytes.Buffer)
	baseCmd.SetOut(buf)
	baseCmd.SetErr(buf)
	wwlog.SetLogWriterErr(buf)
	wwlog.SetLogWriterInfo(buf)
	err := baseCmd.Execute()
	assert.NoError(t, err)

	assert.NotEmpty(t, buf.String(), "output should not be empty")
	assert.Contains(t, buf.String(), content)
}
