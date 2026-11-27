// SPDX-License-Identifier: BSD-3-Clause
// Copyright holders are listed in LICENSE.md.

package show

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func TestShowRender(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	const source = `{{.HostOverlay}} {{basename .BuildSource}} {{.Tags.absent}}`
	tests := []struct {
		name     string
		flags    []string
		template string
		output   string
		err      string
	}{
		{"node", []string{"--render", "node1"}, source, "false test.ww <no value>\n", ""},
		{"host", []string{"--render-host"}, source, "true test.ww <no value>\n", ""},
		{"parse error", []string{"--render-host"}, `{{if}}`, "", "could not parse template"},
		{"execution error", []string{"--render", "node1"}, `partial{{.UnknownField}}`, "", "could not execute template"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env.WriteFile(filepath.Join(testenv.WWOverlaydir, "sample/rootfs/test.ww"), tt.template)
			var output bytes.Buffer
			wwlog.SetLogWriterInfo(&output)
			defer wwlog.SetLogWriterInfo(os.Stdout)

			cmd := GetCommand()
			cmd.SetArgs(append(tt.flags, "--quiet", "sample", "test.ww"))
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			err := cmd.Execute()
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.output, output.String())
		})
	}
}
