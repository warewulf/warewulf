package flags

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_Aliases(t *testing.T) {
	tests := map[string]struct {
		args            []string
		changed         []string
		image           string
		systemOverlays  []string
		runtimeOverlays []string
		deprecated      []string
	}{
		"--container": {
			args:       []string{"--container=img"},
			changed:    []string{"image"},
			image:      "img",
			deprecated: []string{"Flag --container has been deprecated, use --image instead"},
		},
		"-C": {
			args:       []string{"-C", "img"},
			changed:    []string{"image"},
			image:      "img",
			deprecated: []string{"Flag --container has been deprecated, use --image instead"},
		},
		"--wwinit": {
			args:           []string{"--wwinit=a", "--wwinit=b"},
			changed:        []string{"system-overlays"},
			systemOverlays: []string{"a", "b"},
			deprecated:     []string{"Flag --wwinit has been deprecated, use --system-overlays instead"},
		},
		"--wwinit list": {
			args:           []string{"--wwinit=a,b"},
			changed:        []string{"system-overlays"},
			systemOverlays: []string{"a", "b"},
			deprecated:     []string{"Flag --wwinit has been deprecated, use --system-overlays instead"},
		},
		"--runtime": {
			args:            []string{"--runtime=a"},
			changed:         []string{"runtime-overlays"},
			runtimeOverlays: []string{"a"},
			deprecated:      []string{"Flag --runtime has been deprecated, use --runtime-overlays instead"},
		},
		"--system-overlays and --wwinit": {
			args:           []string{"--system-overlays=a", "--wwinit=b"},
			changed:        []string{"system-overlays"},
			systemOverlays: []string{"a", "b"},
			deprecated:     []string{"Flag --wwinit has been deprecated, use --system-overlays instead"},
		},
		"no aliases": {
			args:    []string{"--image=img"},
			changed: []string{"image"},
			image:   "img",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var image string
			var systemOverlays, runtimeOverlays []string
			cmd := &cobra.Command{
				Use:  "test",
				RunE: func(cmd *cobra.Command, args []string) error { return nil },
			}
			cmd.PersistentFlags().StringVar(&image, "image", "", "image name")
			cmd.PersistentFlags().StringSliceVarP(&systemOverlays, "system-overlays", "O", []string{}, "system overlays")
			cmd.PersistentFlags().StringSliceVarP(&runtimeOverlays, "runtime-overlays", "R", []string{}, "runtime overlays")
			AddContainer(cmd)
			AddWwinit(cmd)
			AddRuntime(cmd)
			cmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			assert.NoError(t, cmd.Execute())
			for _, flag := range tt.changed {
				assert.True(t, cmd.Flags().Changed(flag), flag)
			}
			assert.Equal(t, tt.image, image)
			if tt.systemOverlays == nil {
				tt.systemOverlays = []string{}
			}
			if tt.runtimeOverlays == nil {
				tt.runtimeOverlays = []string{}
			}
			assert.Equal(t, tt.systemOverlays, systemOverlays)
			assert.Equal(t, tt.runtimeOverlays, runtimeOverlays)
			for _, msg := range tt.deprecated {
				assert.Contains(t, buf.String(), msg)
			}
			if len(tt.deprecated) == 0 {
				assert.NotContains(t, buf.String(), "deprecated")
			}
		})
	}
}
