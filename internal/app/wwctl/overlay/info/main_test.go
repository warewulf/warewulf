package info

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Overlay_Variables(t *testing.T) {
	tests := []struct {
		name           string
		writeFiles     map[string]string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name: "overlay variables",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/test.ww": `
{{/* .Kernel.Tags.foo: "some help text" */}}
{{/* wwdoc1: First Line */}}
{{ .Node.Tags.bar }}
{{/* wwdoc2: Second Line */}}
`,
			},
			args:        []string{"test-overlay", "test.ww"},
			expectError: false,
			expectedOutput: "First Line\n" +
				"Second Line\n" +
				"\n" +
				"VARIABLE        OPTION  TYPE    HELP\n" +
				"--------        ------  ----    ----\n" +
				".Node.Tags.bar          string  \n",
		},
		{
			name: "overlay variables no file",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/test.ww": ``,
			},
			args:        []string{"test-overlay", "no-file.ww"},
			expectError: true,
		},
		{
			name:        "overlay variables no overlay",
			args:        []string{"no-overlay", "test.ww"},
			expectError: true,
		},
		{
			name: "nested fields multiple levels",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/nested.ww": `
{{ .Kernel.Args }}
{{ .Ipmi.UserName }}
{{ .Ipmi.Ipaddr }}
{{ .Warewulf.Port }}
`,
			},
			args:        []string{"test-overlay", "nested.ww"},
			expectError: false,
			expectedOutput: "VARIABLE        OPTION        TYPE      HELP\n" +
				"--------        ------        ----      ----\n" +
				".Ipmi.Ipaddr    --ipmiaddr    IP        Set the IPMI IP address\n" +
				".Ipmi.UserName  --ipmiuser    string    Set the IPMI username\n" +
				".Kernel.Args    --kernelargs  []string  Set kernel arguments\n" +
				".Warewulf.Port                int       \n",
		},
		{
			name: "template without variables",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/no-vars.ww": `
# Static configuration file
# This file has no template variables
static_value=true
`,
			},
			args:        []string{"test-overlay", "no-vars.ww"},
			expectError: false,
			expectedOutput: `VARIABLE  OPTION  TYPE  HELP
--------  ------  ----  ----
`,
		},
		{
			name: "only documentation no variables",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/doc-only.ww": `
{{/* wwdoc: This is just documentation */}}
{{/* wwdoc2: No actual template variables used */}}
Static content here
`,
			},
			args:        []string{"test-overlay", "doc-only.ww"},
			expectError: false,
			expectedOutput: `This is just documentation
No actual template variables used

VARIABLE  OPTION  TYPE  HELP
--------  ------  ----  ----
`,
		},
		{
			name: "template parse error",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/invalid.ww": `
{{ .Id }
{{ unclosed range .NetDevs }}
`,
			},
			args:        []string{"test-overlay", "invalid.ww"},
			expectError: true,
		},
		{
			name: "inline comment documentation",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/inline-doc.ww": `
{{/* .Id: The unique node identifier */}}
{{/* .Hostname: The node's hostname in DNS */}}
{{/* .Tags.env: Environment tag (prod, dev, test) */}}
{{ .Id }}
{{ .Hostname }}
{{ .Tags.env }}
`,
			},
			args:        []string{"test-overlay", "inline-doc.ww"},
			expectError: false,
			expectedOutput: `VARIABLE   OPTION  TYPE    HELP
--------   ------  ----    ----
.Hostname          string  The node's hostname in DNS
.Id                string  The unique node identifier
.Tags.env          string  Environment tag (prod, dev, test)
`,
		},
		{
			name: "complex conditional branches",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/conditionals.ww": `
{{ if eq .Kernel.Version "5.15" }}
Old kernel
{{ else if gt (len .NetDevs) 0 }}
Has network: {{ .Id }}
{{ else }}
Default: {{ .Hostname }}
{{ end }}
`,
			},
			args:        []string{"test-overlay", "conditionals.ww"},
			expectError: false,
			expectedOutput: "VARIABLE         OPTION           TYPE                     HELP\n" +
				"--------         ------           ----                     ----\n" +
				".Hostname                         string                   \n" +
				".Id                               string                   \n" +
				".Kernel.Version  --kernelversion  string                   Set kernel version\n" +
				".NetDevs                          map[string]*node.NetDev  \n",
		},
		{
			name: "sprig function pipelines",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/sprig.ww": `
{{ .Id | upper }}
{{ .Hostname | lower }}
{{ .Kernel.Args | join " " }}
{{ .Tags.foo | default "bar" }}
`,
			},
			args:        []string{"test-overlay", "sprig.ww"},
			expectError: false,
			expectedOutput: "VARIABLE      OPTION        TYPE      HELP\n" +
				"--------      ------        ----      ----\n" +
				".Hostname                   string    \n" +
				".Id                         string    \n" +
				".Kernel.Args  --kernelargs  []string  Set kernel arguments\n" +
				".Tags.foo                   string    \n",
		},
		{
			name: "include function usage",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/include.ww": `
{{ Include "/etc/passwd" }}
{{ IncludeFrom .ImageName "/etc/group" }}
`,
			},
			args:        []string{"test-overlay", "include.ww"},
			expectError: false,
			expectedOutput: "VARIABLE    OPTION   TYPE    HELP\n" +
				"--------    ------   ----    ----\n" +
				".ImageName  --image  string  Set image name\n",
		},
		{
			name: "mixed documentation types",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/mixed-doc.ww": `
{{/* wwdoc: Configuration for network interfaces */}}
{{/* wwdoc-details: This template generates network configs */}}
{{/* .NetDevs: Map of network devices by name */}}
{{/* .Id: Node identifier */}}
{{ .Id }}
{{ .NetDevs }}
`,
			},
			args:        []string{"test-overlay", "mixed-doc.ww"},
			expectError: false,
			expectedOutput: "Configuration for network interfaces\n" +
				"This template generates network configs\n" +
				"\n" +
				"VARIABLE  OPTION  TYPE                     HELP\n" +
				"--------  ------  ----                     ----\n" +
				".Id               string                   Node identifier\n" +
				".NetDevs          map[string]*node.NetDev  Map of network devices by name\n",
		},
		{
			name: "tag field access",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/tags.ww": `
{{ .Tags.foo }}
{{ .Ipmi.Tags.vlan }}
{{ .NetDevs }}
`,
			},
			args:        []string{"test-overlay", "tags.ww"},
			expectError: false,
			expectedOutput: "VARIABLE         OPTION  TYPE                     HELP\n" +
				"--------         ------  ----                     ----\n" +
				".Ipmi.Tags.vlan          string                   \n" +
				".NetDevs                 map[string]*node.NetDev  \n" +
				".Tags.foo                string                   \n",
		},
		{
			name: "dollar sign root context",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/dollar.ww": `
{{ $.Id }}
{{ $.BuildHost }}
{{ $.Ipaddr }}
`,
			},
			args:        []string{"test-overlay", "dollar.ww"},
			expectError: false,
			expectedOutput: "VARIABLE     OPTION  TYPE    HELP\n" +
				"--------     ------  ----    ----\n" +
				"$.BuildHost          string  \n" +
				"$.Id                 string  \n" +
				"$.Ipaddr             string  \n",
		},
		{
			name: "range over map basic",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/range-map.ww": `
{{/* wwdoc: Network device iteration */}}
{{- range $devname, $netdev := .NetDevs }}
Device: {{ $netdev.Device }}
Type: {{ $netdev.Type }}
IP: {{ $netdev.Ipaddr }}
{{- end }}
`,
			},
			args:        []string{"test-overlay", "range-map.ww"},
			expectError: false,
			expectedOutput: "Network device iteration\n" +
				"\n" +
				"VARIABLE        OPTION    TYPE                     HELP\n" +
				"--------        ------    ----                     ----\n" +
				"$netdev.Device  --netdev  string                   Set the device for given network\n" +
				"$netdev.Ipaddr  --ipaddr  IP                       IPv4 address in given network\n" +
				"$netdev.Type    --type    string                   Set device type of given network\n" +
				".NetDevs                  map[string]*node.NetDev  \n",
		},
		{
			name: "range over map with output verification",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/range-verify.ww": `
{{- range $devname, $netdev := .NetDevs }}
{{ $devname }}
{{ $netdev.Device }}
{{ $netdev.Type }}
{{- end }}
`,
			},
			args:        []string{"test-overlay", "range-verify.ww"},
			expectError: false,
			expectedOutput: "VARIABLE        OPTION    TYPE                     HELP\n" +
				"--------        ------    ----                     ----\n" +
				"$netdev.Device  --netdev  string                   Set the device for given network\n" +
				"$netdev.Type    --type    string                   Set device type of given network\n" +
				".NetDevs                  map[string]*node.NetDev  \n",
		},
		{
			name: "empty collection checks",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/empty.ww": `
{{ if gt (len .NetDevs) 0 }}
Has devices
{{ end }}
{{ if .FileSystems }}
Has filesystems
{{ end }}
`,
			},
			args:        []string{"test-overlay", "empty.ww"},
			expectError: false,
			expectedOutput: "VARIABLE      OPTION  TYPE                         HELP\n" +
				"--------      ------  ----                         ----\n" +
				".FileSystems          map[string]*node.FileSystem  \n" +
				".NetDevs              map[string]*node.NetDev      \n",
		},
		{
			name: "file and abort directives",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/control.ww": `
{{ file "output.conf" }}
{{ if .Tags.enabled }}
Config: {{ .Tags.value }}
{{ else }}
{{ abort }}
{{ end }}
`,
			},
			args:        []string{"test-overlay", "control.ww"},
			expectError: false,
			expectedOutput: "VARIABLE       OPTION  TYPE    HELP\n" +
				"--------       ------  ----    ----\n" +
				".Tags.enabled          string  \n" +
				".Tags.value            string  \n",
		},
		{
			name: "wwdoc comments with trimmed whitespace",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/trimmed-doc.ww": `
{{- /* wwdoc: Configuration for GPU MIG partitions with trimmed whitespace */ -}}
{{- /* wwdoc-details: This template demonstrates wwdoc comments with {{- -}} syntax */ -}}
{{- /* .Tags.gpuMigProfiles: List of MIG profile IDs with GPU indices */ -}}}
GPU Profile: {{ .Tags.gpuMigProfiles }}
`,
			},
			args:        []string{"test-overlay", "trimmed-doc.ww"},
			expectError: false,
			expectedOutput: "Configuration for GPU MIG partitions with trimmed whitespace\n" +
				"This template demonstrates wwdoc comments with {{- -}} syntax\n" +
				"\n" +
				"VARIABLE              OPTION  TYPE    HELP\n" +
				"--------              ------  ----    ----\n" +
				".Tags.gpuMigProfiles          string  List of MIG profile IDs with GPU indices\n",
		},
		{
			name: "wwdoc comments with half-trimmed whitespace",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/half-trimmed-doc.ww": `
{{- /* wwdoc: Configuration for GPU MIG partitions with half-trimmed whitespace */}}
{{- /* wwdoc-details: This template demonstrates wwdoc comments with {{- }} syntax */}}
{{- /* .Tags.gpuMigProfiles: List of MIG profile IDs with GPU indices */}}}
GPU Profile: {{ .Tags.gpuMigProfiles }}
`,
			},
			args:        []string{"test-overlay", "half-trimmed-doc.ww"},
			expectError: false,
			expectedOutput: "Configuration for GPU MIG partitions with half-trimmed whitespace\n" +
				"This template demonstrates wwdoc comments with {{- }} syntax\n" +
				"\n" +
				"VARIABLE              OPTION  TYPE    HELP\n" +
				"--------              ------  ----    ----\n" +
				".Tags.gpuMigProfiles          string  List of MIG profile IDs with GPU indices\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			warewulfd.SetNoDaemon()

			for path, content := range tt.writeFiles {
				env.WriteFile(path, content)
			}
			baseCmd := GetCommand()
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)

			baseCmd.SetArgs(tt.args)
			err := baseCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedOutput != "" {
				output := buf.String()
				assert.Equal(t, tt.expectedOutput, output)
			} else {
				// For tests without expected output, print actual output for debugging
				// This helps understand what variables are being detected
				if testing.Verbose() {
					t.Logf("Output:\n%s", buf.String())
				}
			}
		})
	}
}
