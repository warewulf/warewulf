package variables

import (
	"bytes"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Overlay_Variables(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	conf := env.Configure()

	// be quiet
	wwlog.SetLogFormatter(func(int, *wwlog.LogRecord) string { return "" })

	overlayDir := path.Join(conf.Paths.SiteOverlaydir(), "test-overlay")
	err := os.MkdirAll(overlayDir, 0755)
	if err != nil {
		t.Fatalf("could not create overlay dir: %v", err)
	}

	templateContent := `
{{ .Kernel.Tags.foo }}
{{ .Node.Tags.bar }}
{{ .Cluster.Tags.baz }}
`
	templatePath := path.Join(overlayDir, "test.ww")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("could not write template file: %v", err)
	}

	// Redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	baseCmd.SetArgs([]string{"test-overlay", "test.ww"})
	err = baseCmd.Execute()
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("could not read stdout: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	expected := []string{
		".Kernel.Tags.foo",
		".Node.Tags.bar",
		".Cluster.Tags.baz",
	}

	outputLines := strings.Split(output, "\n")
	assert.ElementsMatch(t, expected, outputLines)
}

func Test_Overlay_Variables_No_File(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	conf := env.Configure()

	// be quiet
	wwlog.SetLogFormatter(func(int, *wwlog.LogRecord) string { return "" })
	overlayDir := path.Join(conf.Paths.SiteOverlaydir(), "test-overlay")
	err := os.MkdirAll(overlayDir, 0755)
	if err != nil {
		t.Fatalf("could not create overlay dir: %v", err)
	}

	baseCmd.SetArgs([]string{"test-overlay", "test.ww"})
	err = baseCmd.Execute()
	assert.Error(t, err)
}

func Test_Overlay_Variables_No_Overlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.Configure()

	// be quiet
	wwlog.SetLogFormatter(func(int, *wwlog.LogRecord) string { return "" })

	baseCmd.SetArgs([]string{"no-overlay", "test.ww"})
	err := baseCmd.Execute()
	assert.Error(t, err)
}
