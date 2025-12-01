package keys

import (
	"bytes"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Keys(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// Define a keystore path within the test environment
	keystoreRelPath := "etc/warewulf/keys"
	keystorePath := env.GetPath(keystoreRelPath)

	// Reload configuration to pick up the change
	env.Configure()

	t.Run("keys create", func(t *testing.T) {
		baseCmd := GetCommand()
		// Reset flags
		create = true
		importPath = ""
		exportPath = ""

		baseCmd.SetArgs([]string{"--create"})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)

		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.FileExists(t, path.Join(keystorePath, "warewulf.key"))
		assert.FileExists(t, path.Join(keystorePath, "warewulf.crt"))
		assert.FileExists(t, path.Join(keystorePath, "warewulf.rsa.pub"))
	})

	t.Run("keys exist check", func(t *testing.T) {
		baseCmd := GetCommand()
		// Reset flags
		create = true // Even with create, it should say they exist
		importPath = ""
		exportPath = ""

		baseCmd.SetArgs([]string{"--create"})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)

		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Keys already exist")
	})

	t.Run("keys display", func(t *testing.T) {
		baseCmd := GetCommand()
		// Reset flags
		create = false
		importPath = ""
		exportPath = ""

		baseCmd.SetArgs([]string{})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)

		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Private Key:")
		assert.Contains(t, buf.String(), "Certificate:")
		assert.Contains(t, buf.String(), "Key Size:")
		assert.Contains(t, buf.String(), "Issuer:")
		assert.Contains(t, buf.String(), "Subject:")
		assert.Contains(t, buf.String(), "Valid From:")
		assert.Contains(t, buf.String(), "Valid Until:")
	})

	t.Run("keys export", func(t *testing.T) {
		exportRelPath := "exported_keys"
		exportFullPath := env.GetPath(exportRelPath)

		baseCmd := GetCommand()
		// Reset flags
		create = false
		importPath = ""
		exportPath = exportFullPath

		baseCmd.SetArgs([]string{"--export", exportFullPath})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)

		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.FileExists(t, path.Join(exportFullPath, "warewulf.key"))
		assert.FileExists(t, path.Join(exportFullPath, "warewulf.crt"))
		assert.Contains(t, buf.String(), "Exported keys to")
	})
}
