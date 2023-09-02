package overlay

import (
	"io"
	"os"
	"path"
	"strings"
	"sort"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sassoftware/go-rpmutils/cpio"
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
)

var buildOverlayTests = []struct{
	description string
	nodeName string
	context string
	overlays []string
	image string
	contents []string
}{
	{"empty", "", "", nil, "", nil},
	{"empty node", "node1", "", nil, "", nil},
	{"empty context", "", "system", nil, "", nil},
	{"empty overlay", "", "", []string{"o1"}, "o1.img", []string{"o1.txt"}},
	{"single overlay", "node1", "", []string{"o1"}, "node1/o1.img", []string{"o1.txt"}},
	{"multiple overlays", "node1", "", []string{"o1", "o2"}, "node1/o1-o2.img", []string{"o1.txt", "o2.txt"}},
	{"empty system overlay", "node1", "system", nil, "", nil},
	{"empty runtime overlay", "node1", "runtime", nil, "", nil},
	{"single system overlay", "node1", "system", []string{"o1"}, "node1/__SYSTEM__.img", []string{"o1.txt"}},
	{"single runtime overlay", "node1", "runtime", []string{"o1"}, "node1/__RUNTIME__.img", []string{"o1.txt"}},
	{"two system overlays", "node1", "system", []string{"o1", "o2"}, "node1/__SYSTEM__.img", []string{"o1.txt", "o2.txt"}},
	{"two runtime overlays", "node1", "runtime", []string{"o1", "o2"}, "node1/__RUNTIME__.img", []string{"o1.txt", "o2.txt"}},
}

func Test_BuildOverlay(t *testing.T) {
	conf := warewulfconf.Get()
	overlayDir, overlayDirErr := os.MkdirTemp(os.TempDir(), "ww-test-overlay-*")
	assert.NoError(t, overlayDirErr)
	defer os.RemoveAll(overlayDir)
	conf.Paths.WWOverlaydir = overlayDir
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o1"), 0700))
	{ _, err := os.Create(path.Join(overlayDir, "o1", "o1.txt")); assert.NoError(t, err) }
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o2"), 0700))
	{ _, err := os.Create(path.Join(overlayDir, "o2", "o2.txt")); assert.NoError(t, err) }

	for _, tt := range buildOverlayTests {
		assert.True(t, (tt.image != "" && tt.contents != nil) || (tt.image == "" && tt.contents == nil),
			"image and contents must eiher be populated or empty together")

		nodeInfo := node.NodeInfo{}
		nodeInfo.Id.Set(tt.nodeName)
		t.Run(tt.description, func(t *testing.T) {
			provisionDir, provisionDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
			assert.NoError(t, provisionDirErr)
			defer os.RemoveAll(provisionDir)
			conf.Paths.WWProvisiondir = provisionDir

			err := BuildOverlay(nodeInfo, tt.context, tt.overlays)
			if len(tt.image) > 0 {
				image := path.Join(provisionDir, "overlays", tt.image)
				assert.FileExists(t, image)
				assert.NoError(t, err)

				sort.Strings(tt.contents)
				files := cpioFiles(t, image)
				sort.Strings(files)
				assert.Equal(t, tt.contents, files)
			} else {
				assert.Error(t, err)
				dirName := path.Join(provisionDir, "overlays", tt.nodeName)
				isEmpty := dirIsEmpty(t, dirName)
				assert.True(t, isEmpty, "%v should be empty, but isn't", dirName)
			}
		})
	}
}

var buildAllOverlaysTests = []struct{
	description string
	nodes []string
	systemOverlays []string
	runtimeOverlays []string
	succeed bool
}{
	{"no nodes", nil, nil, nil, true},
	{"single empty node", []string{"node1"}, nil, nil, false},
	{"two empty node", []string{"node1", "node2"}, nil, nil, false},
	{"single node with system overlay", []string{"node1"},
		[]string{"o1"}, nil, false},
	{"two nodes with system overlays", []string{"node1", "node2"},
		[]string{"o1", "o1,o2"}, nil, false},
	{"single node with runtime overlay", []string{"node1"},
		nil, []string{"o1"}, false},
	{"two nodes with runtime overlays", []string{"node1", "node2"},
		nil, []string{"o1", "o1,o2"}, false},
	{"stingle node with full overlays", []string{"node1"},
		[]string{"o1"}, []string{"o2"}, true},
	{"two nodes with full overlays", []string{"node1", "node2"},
		[]string{"o1", "o1,o2"}, []string{"o2", "o2"}, true},
}


func Test_BuildAllOverlays(t *testing.T) {
	conf := warewulfconf.Get()
	overlayDir, overlayDirErr := os.MkdirTemp(os.TempDir(), "ww-test-overlay-*")
	assert.NoError(t, overlayDirErr)
	defer os.RemoveAll(overlayDir)
	conf.Paths.WWOverlaydir = overlayDir
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o1"), 0700))
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o2"), 0700))

	for _, tt := range buildAllOverlaysTests {
		t.Run(tt.description, func(t *testing.T) {
			provisionDir, provisionDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
			assert.NoError(t, provisionDirErr)
			defer os.RemoveAll(provisionDir)
			conf.Paths.WWProvisiondir = provisionDir

			var nodes []node.NodeInfo
			for i, nodeName := range tt.nodes {
				nodeInfo := node.NodeInfo{}
				nodeInfo.Id.Set(nodeName)
				if tt.systemOverlays != nil {
					nodeInfo.SystemOverlay.SetSlice(strings.Split(tt.systemOverlays[i], ","))
				}
				if tt.runtimeOverlays != nil {
					nodeInfo.RuntimeOverlay.SetSlice(strings.Split(tt.runtimeOverlays[i], ","))
				}
				nodes = append(nodes, nodeInfo)
			}
			err := BuildAllOverlays(nodes)
			if !tt.succeed {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, nodeName := range tt.nodes {
					assert.FileExists(t, path.Join(provisionDir, "overlays", nodeName, "__SYSTEM__.img"))
					assert.FileExists(t, path.Join(provisionDir, "overlays", nodeName, "__RUNTIME__.img"))
				}
			}
		})
	}
}

var buildSpecificOverlaysTests = []struct{
	description string
	nodes []string
	overlays string
	images []string
	succeed bool
}{
	{"no nodes", nil, "", nil, true},
	{"single empty node", []string{"node1"}, "", nil, false},
	{"two empty node", []string{"node1", "node2"}, "", nil, false},
	{"single node with single overlay", []string{"node1"}, "o1",
		[]string{"node1/o1.img"}, true},
	{"two nodes with single overlay", []string{"node1", "node2"}, "o1",
		[]string{"node1/o1.img", "node2/o1.img"}, true},
	{"single node with multi overlay", []string{"node1"}, "o1,o2",
		[]string{"node1/o1.img", "node1/o2.img"}, true},
	{"two nodes with multi overlays", []string{"node1", "node2"}, "o1,o2",
		[]string{"node1/o1.img", "node1/o2.img", "node2/o1.img", "node2/o2.img"}, true},
}


func Test_BuildSpecificOverlays(t *testing.T) {
	conf := warewulfconf.Get()
	overlayDir, overlayDirErr := os.MkdirTemp(os.TempDir(), "ww-test-overlay-*")
	assert.NoError(t, overlayDirErr)
	defer os.RemoveAll(overlayDir)
	conf.Paths.WWOverlaydir = overlayDir
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o1"), 0700))
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o2"), 0700))

	for _, tt := range buildSpecificOverlaysTests {
		t.Run(tt.description, func(t *testing.T) {
			provisionDir, provisionDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
			assert.NoError(t, provisionDirErr)
			defer os.RemoveAll(provisionDir)
			conf.Paths.WWProvisiondir = provisionDir

			var nodes []node.NodeInfo
			for _, nodeName := range tt.nodes {
				nodeInfo := node.NodeInfo{}
				nodeInfo.Id.Set(nodeName)
				nodes = append(nodes, nodeInfo)
			}
			err := BuildSpecificOverlays(nodes, strings.Split(tt.overlays, ","))
			if !tt.succeed {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, image := range tt.images {
					assert.FileExists(t, path.Join(provisionDir, "overlays", image))
				}
			}
		})
	}
}

func dirIsEmpty(t *testing.T, name string) bool {
	f, err := os.Open(name)
	if err != nil {
		t.Log(err)
		return true
	}
	defer f.Close()

	dirnames, err2 := f.Readdirnames(1)
	if err2 == io.EOF {
		t.Log(err2)
		return true
	}
	t.Log(dirnames)
	return false
}

func cpioFiles(t *testing.T, name string) (files []string) {
	f, openErr := os.Open(name)
	if openErr != nil { return }
	defer f.Close()

	reader := cpio.NewReader(f)
	for {
		header, err := reader.Next()
		if err != nil { return }
		files = append(files, header.Filename())
	}
}
