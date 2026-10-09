package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestSecureRoutes(t *testing.T) {
	tests := map[string]struct {
		input   string
		runtime bool
		system  bool
		files   bool
		output  string
		err     string
	}{
		"unset":      {input: `port: 9873`, runtime: true, files: true, output: `true`},
		"null":       {input: `secure: ~`, runtime: true, files: true, output: `null`},
		"true":       {input: `secure: true`, runtime: true, files: true, output: `true`},
		"false":      {input: `secure: false`, output: `false`},
		"yes":        {input: `secure: yes`, runtime: true, files: true, output: `true`},
		"off":        {input: `secure: off`, output: `false`},
		"all":        {input: `secure: all`, runtime: true, system: true, files: true, output: `all`},
		"empty list": {input: `secure: []`, output: `[]`},
		"list": {
			input:  "secure:\n  - files\n  - system\n  - runtime\n  - system",
			system: true, runtime: true, files: true,
			output: "- runtime\n- system\n- files",
		},
		"system only":   {input: `secure: [system]`, system: true, output: `- system`},
		"unknown entry": {input: `secure: [image]`, err: `entry "image"`},
		"null entry":    {input: `secure: [runtime, ~]`, err: `list at line 2`},
		"blank entry":   {input: "secure:\n  - ", err: `list at line 3`},
		"nested list":   {input: `secure: [[runtime]]`, err: `list at line 2`},
		"bad scalar":    {input: `secure: sometimes`, err: `"sometimes"`},
		"mapping":       {input: `secure: {runtime: true}`, err: `must be true, false, all, or a list of: runtime, system, files`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			conf := New()
			err := conf.Parse([]byte("warewulf:\n  "+strings.ReplaceAll(tt.input, "\n", "\n  ")), false)
			if tt.err != "" {
				assert.ErrorContains(t, err, tt.err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.runtime, conf.Warewulf.SecureRuntime())
			assert.Equal(t, tt.runtime, conf.Warewulf.Secure())
			assert.Equal(t, tt.system, conf.Warewulf.SecureSystem())
			assert.Equal(t, tt.files, conf.Warewulf.SecureFiles())

			out, err := yaml.Marshal(conf.Warewulf.SecureP)
			assert.NoError(t, err)
			assert.Equal(t, tt.output, strings.TrimSpace(string(out)))
		})
	}
}

func TestSecureFilesOverride(t *testing.T) {
	tests := map[string]struct {
		input   string
		runtime bool
		files   bool
	}{
		"true, files false":   {"secure: true\nsecure files: false", true, false},
		"false, files true":   {"secure: false\nsecure files: true", false, true},
		"unset, files false":  {"secure files: false", true, false},
		"list, files false":   {"secure: [files]\nsecure files: false", false, false},
		"list, files true":    {"secure: [runtime]\nsecure files: true", true, true},
		"all, files false":    {"secure: all\nsecure files: false", true, false},
		"empty, files true":   {"secure: []\nsecure files: true", false, true},
		"system, files false": {"secure: [system]\nsecure files: false", false, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			conf := New()
			assert.NoError(t, conf.Parse([]byte("warewulf:\n  "+strings.ReplaceAll(tt.input, "\n", "\n  ")), false))
			assert.Equal(t, tt.runtime, conf.Warewulf.SecureRuntime())
			assert.Equal(t, tt.files, conf.Warewulf.SecureFiles())
		})
	}
}

func TestSecureRoutesWith(t *testing.T) {
	tests := map[string]struct {
		secure  *SecureRoutes
		enabled bool
		output  string
	}{
		"nil agrees":       {nil, true, `null`},
		"nil disagrees":    {nil, false, `- runtime`},
		"false agrees":     {&SecureRoutes{form: secureFormBool}, false, `false`},
		"all disagrees":    {&SecureRoutes{form: secureFormAll}, false, "- runtime\n- system"},
		"list agrees":      {NewSecureRoutes(SecureRouteSystem), false, `- system`},
		"list disagrees":   {NewSecureRoutes(SecureRouteSystem), true, "- system\n- files"},
		"files to nothing": {NewSecureRoutes(SecureRouteFiles), false, `[]`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			out, err := yaml.Marshal(tt.secure.With(SecureRouteFiles, tt.enabled))
			assert.NoError(t, err)
			assert.Equal(t, tt.output, strings.TrimSpace(string(out)))
		})
	}
}
