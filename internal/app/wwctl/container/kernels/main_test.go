package kernels

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_List(t *testing.T) {
	tests := map[string]struct {
		files  map[string][]string
		args   []string
		stdout string
	}{
		"default": {
			files: map[string][]string{},
			args:  []string{},
			stdout: `
Container  Kernel  Version  Preferred  Nodes
---------  ------  -------  ---------  -----
`,
		},
		"list": {
			files: map[string][]string{
				"container1": []string{
					"/boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64",
					"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
					"/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
				},
				"container2": []string{
					"/boot/vmlinuz-0-rescue-eb46964329b146e39518c625feab3ea0",
					"/boot/vmlinuz-5.14.0-362.24.1.el9_3.aarch64",
					"/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64+debug",
					"/boot/vmlinuz-5.14.0-284.30.1.el9_2.aarch64",
					"/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64",
				},
			},
			args: []string{},
			stdout: `
Container   Kernel                                                   Version          Preferred  Nodes
---------   ------                                                   -------          ---------  -----
container1  /boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64               4.14.0-427.18.1  false      0
container1  /boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64               5.14.0-427.18.1  false      0
container1  /boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64               5.14.0-427.24.1  true       0
container2  /boot/vmlinuz-0-rescue-eb46964329b146e39518c625feab3ea0  --               false      0
container2  /boot/vmlinuz-5.14.0-284.30.1.el9_2.aarch64              5.14.0-284.30.1  false      0
container2  /boot/vmlinuz-5.14.0-362.24.1.el9_3.aarch64              5.14.0-362.24.1  false      0
container2  /boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64              5.14.0-427.31.1  true       0
container2  /boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64+debug        5.14.0-427.31.1  false      0
`,
		},
		"single container": {
			files: map[string][]string{
				"container1": []string{
					"/boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64",
					"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
					"/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
				},
				"container2": []string{
					"/boot/vmlinuz-0-rescue-eb46964329b146e39518c625feab3ea0",
					"/boot/vmlinuz-5.14.0-362.24.1.el9_3.aarch64",
					"/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64+debug",
					"/boot/vmlinuz-5.14.0-284.30.1.el9_2.aarch64",
					"/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64",
				},
			},
			args: []string{"container2"},
			stdout: `
Container   Kernel                                                   Version          Preferred  Nodes
---------   ------                                                   -------          ---------  -----
container2  /boot/vmlinuz-0-rescue-eb46964329b146e39518c625feab3ea0  --               false      0
container2  /boot/vmlinuz-5.14.0-284.30.1.el9_2.aarch64              5.14.0-284.30.1  false      0
container2  /boot/vmlinuz-5.14.0-362.24.1.el9_3.aarch64              5.14.0-362.24.1  false      0
container2  /boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64              5.14.0-427.31.1  true       0
container2  /boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64+debug        5.14.0-427.31.1  false      0
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll(t)
			for container, files := range tt.files {
				rootfs := filepath.Join(filepath.Join("/var/lib/warewulf/chroots", container), "rootfs")
				for _, file := range files {
					env.CreateFile(t, filepath.Join(rootfs, file))
				}
			}
			buf := new(bytes.Buffer)
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tt.stdout), strings.TrimSpace(buf.String()))
		})
	}
}
