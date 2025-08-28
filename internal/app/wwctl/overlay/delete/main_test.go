package delete

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_OverlayDelete(t *testing.T) {
	tests := []struct {
		name        string
		overlayName string
		fileName    string
		setup       func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string)
		wantErr     bool
		verify      func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string)
	}{
		{
			name:        "delete site overlay",
			overlayName: "test-overlay",
			setup: func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {
				siteOverlayDir := filepath.Join(conf.Paths.SiteOverlaydir(), overlayName)
				err := os.MkdirAll(siteOverlayDir, 0755)
				assert.NoError(t, err)
			},
			wantErr: false,
			verify: func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {
				siteOverlayDir := filepath.Join(conf.Paths.SiteOverlaydir(), overlayName)
				_, err := os.Stat(siteOverlayDir)
				assert.True(t, os.IsNotExist(err), "site overlay directory should have been deleted")
			},
		},
		{
			name:        "delete file in site overlay",
			overlayName: "test-overlay-file",
			fileName:    "test.txt",
			setup: func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {
				siteOverlayDir := filepath.Join(conf.Paths.SiteOverlaydir(), overlayName)
				err := os.MkdirAll(siteOverlayDir, 0755)
				assert.NoError(t, err)
				err = os.WriteFile(filepath.Join(siteOverlayDir, fileName), []byte("test"), 0644)
				assert.NoError(t, err)
			},
			wantErr: false,
			verify: func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {
				siteOverlayDir := filepath.Join(conf.Paths.SiteOverlaydir(), overlayName)
				_, err := os.Stat(filepath.Join(siteOverlayDir, fileName))
				assert.True(t, os.IsNotExist(err), "file in site overlay should have been deleted")
				_, err = os.Stat(siteOverlayDir)
				assert.NoError(t, err, "site overlay directory should still exist")
			},
		},
		{
			name:        "delete non-existent overlay",
			overlayName: "non-existent-overlay",
			setup:       func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {},
			wantErr:     true,
			verify:      func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {},
		},
		{
			name:        "delete distribution overlay",
			overlayName: "dist-overlay",
			setup: func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {
				distOverlayDir := filepath.Join(conf.Paths.DistributionOverlaydir(), overlayName)
				err := os.MkdirAll(distOverlayDir, 0755)
				assert.NoError(t, err)
			},
			wantErr: true,
			verify: func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {
				distOverlayDir := filepath.Join(conf.Paths.DistributionOverlaydir(), overlayName)
				_, err := os.Stat(distOverlayDir)
				assert.NoError(t, err, "distribution overlay directory should not have been deleted")
			},
		},
		{
			name:        "delete file in distribution overlay",
			overlayName: "dist-overlay-file",
			fileName:    "test.txt",
			setup: func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {
				distOverlayDir := filepath.Join(conf.Paths.DistributionOverlaydir(), overlayName)
				err := os.MkdirAll(distOverlayDir, 0755)
				assert.NoError(t, err)
				err = os.WriteFile(filepath.Join(distOverlayDir, fileName), []byte("test"), 0644)
				assert.NoError(t, err)
			},
			// This will trigger a copy-on-write. A site overlay will be created.
			wantErr: false,
			verify: func(t *testing.T, conf *config.WarewulfYaml, overlayName string, fileName string) {
				distOverlayDir := filepath.Join(conf.Paths.DistributionOverlaydir(), overlayName)
				_, err := os.Stat(filepath.Join(distOverlayDir, fileName))
				assert.NoError(t, err, "file in distribution overlay should still exist")

				siteOverlayDir := filepath.Join(conf.Paths.SiteOverlaydir(), overlayName)
				_, err = os.Stat(siteOverlayDir)
				assert.NoError(t, err, "site overlay should have been created")

				_, err = os.Stat(filepath.Join(siteOverlayDir, fileName))
				assert.True(t, os.IsNotExist(err), "file should not exist in the new site overlay")
			},
		},
	}

	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			conf := env.Configure()

			tt.setup(t, conf, tt.overlayName, tt.fileName)

			var args []string
			if tt.fileName != "" {
				args = []string{tt.overlayName, tt.fileName, "--force"}
			} else {
				args = []string{tt.overlayName, "--force"}
			}

			baseCmd := GetCommand()
			baseCmd.SetArgs(args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err := baseCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.verify(t, conf, tt.overlayName, tt.fileName)
		})
	}
}
