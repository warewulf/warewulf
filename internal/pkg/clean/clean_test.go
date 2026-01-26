package clean

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_CleanOverlays(t *testing.T) {
	tests := map[string]struct {
		nodesConf        string   // nodes.conf content
		overlayDirs      []string // directories to create under overlays/
		overlayFiles     []string // files to create under overlays/ (should be skipped)
		skipOverlaySetup bool     // skip creating overlay directory structure
		wantPreserved    []string // overlay dirs that should exist after cleanup
		wantDeleted      []string // overlay dirs that should not exist after cleanup
		wantErr          bool     // expect error from CleanOverlays
	}{
		"preserves valid node, deletes orphaned node": {
			nodesConf: `nodeprofiles: {}
nodes:
  node1: {}`,
			overlayDirs:   []string{"node1", "node2"},
			wantPreserved: []string{"node1"},
			wantDeleted:   []string{"node2"},
			wantErr:       false,
		},
		"empty overlay directory": {
			nodesConf: `nodeprofiles: {}
nodes:
  node1: {}`,
			overlayDirs:   []string{},
			wantPreserved: []string{},
			wantDeleted:   []string{},
			wantErr:       false,
		},
		"skips files in overlay directory": {
			nodesConf: `nodeprofiles: {}
nodes:
  node1: {}`,
			overlayDirs:   []string{"node1"},
			overlayFiles:  []string{"somefile.txt", "another.img"},
			wantPreserved: []string{"node1"},
			wantDeleted:   []string{},
			wantErr:       false,
		},
		"multiple orphaned nodes deleted": {
			nodesConf: `nodeprofiles: {}
nodes:
  node1: {}`,
			overlayDirs:   []string{"node1", "node2", "node3", "node4"},
			wantPreserved: []string{"node1"},
			wantDeleted:   []string{"node2", "node3", "node4"},
			wantErr:       false,
		},
		"all nodes preserved when all exist in database": {
			nodesConf: `nodeprofiles: {}
nodes:
  node1: {}
  node2: {}
  node3: {}`,
			overlayDirs:   []string{"node1", "node2", "node3"},
			wantPreserved: []string{"node1", "node2", "node3"},
			wantDeleted:   []string{},
			wantErr:       false,
		},
		"empty node database deletes all overlays": {
			nodesConf: `nodeprofiles: {}
nodes: {}`,
			overlayDirs:   []string{"node1", "node2"},
			wantPreserved: []string{},
			wantDeleted:   []string{"node1", "node2"},
			wantErr:       false,
		},
		"error when nodes.conf is invalid": {
			nodesConf:   `this is not valid YAML: [[[`,
			overlayDirs: []string{"node1"},
			wantErr:     true,
		},
		"error when overlay directory is missing": {
			nodesConf: `nodeprofiles: {}
nodes:
  node1: {}`,
			skipOverlaySetup: true,
			wantErr:          true,
		},
		"skips directories with suspicious names starting with ..": {
			nodesConf: `nodeprofiles: {}
nodes:
  node1: {}`,
			overlayDirs:   []string{"node1", "..suspicious", "..hidden"},
			wantPreserved: []string{"node1"},
			// Directories starting with ".." are skipped by path traversal protection
			wantDeleted: []string{},
			wantErr:     false,
		},
		"handles directory names with dots correctly": {
			nodesConf: `nodeprofiles: {}
nodes:
  node1: {}
  node.valid: {}`,
			overlayDirs:   []string{"node1", "node.valid", "node.orphaned"},
			wantPreserved: []string{"node1", "node.valid"},
			wantDeleted:   []string{"node.orphaned"},
			wantErr:       false,
		},
		"validates path traversal protection with edge case names": {
			nodesConf: `nodeprofiles: {}
nodes:
  validnode: {}`,
			// Test various edge cases that should be handled by validation
			overlayDirs:   []string{"validnode", "..test", "...triple"},
			wantPreserved: []string{"validnode"},
			// Names starting with ".." (including "...triple") are skipped by path traversal protection
			wantDeleted: []string{},
			wantErr:     false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()

			// Setup nodes.conf
			if tt.nodesConf != "" {
				env.WriteFile("etc/warewulf/nodes.conf", tt.nodesConf)
			}

			// Setup overlay directory structure (unless explicitly skipped)
			if !tt.skipOverlaySetup {
				// Ensure the overlay directory exists (even if empty)
				env.MkdirAll("srv/warewulf/overlays")

				// Setup overlay directories
				for _, dir := range tt.overlayDirs {
					// Create a file inside each directory to make it a proper overlay dir
					env.WriteFile(filepath.Join("srv/warewulf/overlays", dir, "__SYSTEM__.img"), "Fake System")
				}

				// Setup overlay files (non-directories)
				for _, file := range tt.overlayFiles {
					env.WriteFile(filepath.Join("srv/warewulf/overlays", file), "test content")
				}
			} else {
				// For tests that skip setup, remove the overlay directory
				overlayPath := env.GetPath("srv/warewulf/overlays")
				os.RemoveAll(overlayPath)
			}

			// Execute
			err := CleanOverlays()

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify preserved directories exist
				for _, dir := range tt.wantPreserved {
					dirPath := env.GetPath(filepath.Join("srv/warewulf/overlays", dir))
					assert.DirExists(t, dirPath, "expected directory %s to be preserved", dir)
				}

				// Verify deleted directories do not exist
				for _, dir := range tt.wantDeleted {
					dirPath := env.GetPath(filepath.Join("srv/warewulf/overlays", dir))
					assert.NoDirExists(t, dirPath, "expected directory %s to be deleted", dir)
				}

				// Verify files were not deleted (when specified)
				for _, file := range tt.overlayFiles {
					filePath := env.GetPath(filepath.Join("srv/warewulf/overlays", file))
					assert.FileExists(t, filePath, "expected file %s to be skipped", file)
				}
			}
		})
	}
}
