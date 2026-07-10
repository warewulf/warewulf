package edit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_Node_Edit(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// 1. Initial nodes.conf database with comments
	initialDB := `
nodeprofiles:
  default:
    comment: This profile is automatically included for each node
nodes:
  # Original comment for node1
  node1:
    comment: original comment
`
	env.WriteFile("etc/warewulf/nodes.conf", initialDB)
	warewulfd.SetNoDaemon()

	// 2. Set up mock editor
	mockEditorPath := filepath.Join(t.TempDir(), "mock-editor.sh")
	mockEditorContent := `#!/bin/sh
# Keep the header up to and including the separator, then append modified nodes with comments
sed -i '/# DO NOT EDIT ABOVE THIS LINE, CHANGES WILL BE LOST/q' "$1"
cat << 'EOF' >> "$1"
node1:
  # This is a new comment from the editor for node1
  comment: edited comment
EOF
`
	err := os.WriteFile(mockEditorPath, []byte(mockEditorContent), 0o755)
	assert.NoError(t, err)
	_ = os.Setenv("EDITOR", mockEditorPath)
	defer func() {
		_ = os.Unsetenv("EDITOR")
	}()
	// 3. Run the node edit command with --yes
	baseCmd := GetCommand()
	baseCmd.SetArgs([]string{"--yes", "node1"})
	buf := new(bytes.Buffer)
	baseCmd.SetOut(buf)
	baseCmd.SetErr(buf)

	err = baseCmd.Execute()
	t.Logf("Buffer output:\n%s", buf.String())
	assert.NoError(t, err)

	// 4. Verify the persisted nodes.conf content
	content := env.ReadFile("etc/warewulf/nodes.conf")
	t.Logf("Resulting nodes.conf:\n%s", content)

	// Confirm values are updated and comments are preserved/updated
	assert.True(t, strings.Contains(content, "comment: edited comment"))
	assert.True(t, strings.Contains(content, "# This is a new comment from the editor for node1"))
}
