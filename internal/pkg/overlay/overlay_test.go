package overlay

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/util"

	"github.com/cavaliergopher/cpio"
)

func Test_FindOverlays(t *testing.T) {
	tests := map[string]struct {
		distOverlays []string
		siteOverlays []string
		overlayList  []string
	}{
		"dist overlays": {
			distOverlays: []string{"do1", "do2", "do3"},
			overlayList:  []string{"do1", "do2", "do3"},
		},
		"site overlays": {
			siteOverlays: []string{"so1", "so2", "so3"},
			overlayList:  []string{"so1", "so2", "so3"},
		},
		"both overlays": {
			distOverlays: []string{"do1", "do2"},
			siteOverlays: []string{"so3"},
			overlayList:  []string{"do1", "do2", "so3"},
		},
		"shadowed overlay": {
			distOverlays: []string{"do1", "o1"},
			siteOverlays: []string{"o1", "so1"},
			overlayList:  []string{"do1", "o1", "so1"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			for _, overlay := range tt.distOverlays {
				env.MkdirAll(filepath.Join("usr/share/warewulf/overlays", overlay))
			}
			for _, overlay := range tt.siteOverlays {
				env.MkdirAll(filepath.Join("var/lib/warewulf/overlays", overlay))
			}
			overlayList := FindOverlays()
			assert.Equal(t, tt.overlayList, overlayList)
		})
	}
}

func Test_OverlayMethods(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// Setup test data
	sitedir := "/var/lib/warewulf/overlays/"
	distdir := "/usr/share/warewulf/overlays/"

	env.WriteFile(path.Join(sitedir, "siteonly/rootfs/testfile"), "a site overlay")
	env.WriteFile(path.Join(distdir, "distonly/rootfs/testfile"), "a distribution overlay")
	env.WriteFile(path.Join(sitedir, "legacy/testfile"), "a legacy overlay")
	env.WriteFile(path.Join(sitedir, "both/rootfs/testfile"), "the site version")
	env.WriteFile(path.Join(distdir, "both/rootfs/testfile"), "the distribution version")

	tests := map[string]struct {
		name    string
		path    string
		rootfs  string
		file    string
		content string
		exists  bool
		isSite  bool
		isDist  bool
	}{
		"site overlay": {
			name:    "siteonly",
			path:    path.Join(sitedir, "siteonly"),
			rootfs:  path.Join(sitedir, "siteonly/rootfs"),
			file:    path.Join(sitedir, "siteonly/rootfs/testfile"),
			content: "a site overlay",
			exists:  true,
			isSite:  true,
			isDist:  false,
		},
		"distribution overlay": {
			name:    "distonly",
			path:    path.Join(distdir, "distonly"),
			rootfs:  path.Join(distdir, "distonly/rootfs"),
			file:    path.Join(distdir, "distonly/rootfs/testfile"),
			content: "a distribution overlay",
			exists:  true,
			isSite:  false,
			isDist:  true,
		},
		"overlapping overlay": {
			name:    "both",
			path:    path.Join(sitedir, "both"),
			rootfs:  path.Join(sitedir, "both/rootfs"),
			file:    path.Join(sitedir, "both/rootfs/testfile"),
			content: "the site version",
			exists:  true,
			isSite:  true,
			isDist:  false,
		},
		"legacy overlay": {
			name:    "legacy",
			path:    path.Join(sitedir, "legacy"),
			rootfs:  path.Join(sitedir, "legacy"),
			file:    path.Join(sitedir, "legacy/testfile"),
			content: "",
			exists:  true,
			isSite:  true,
			isDist:  false,
		},
		"missing overlay": {
			name:    "absent",
			path:    path.Join(sitedir, "absent"),
			rootfs:  path.Join(sitedir, "absent/rootfs"),
			file:    path.Join(sitedir, "absent/rootfs/testfile"),
			content: "",
			exists:  false,
			isSite:  true,
			isDist:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			overlay := GetOverlay(tt.name)
			assert.Equal(t, tt.name, overlay.Name())
			assert.Equal(t, env.GetPath(tt.path), overlay.Path())
			assert.Equal(t, env.GetPath(tt.rootfs), overlay.Rootfs())
			assert.Equal(t, env.GetPath(tt.file), overlay.File("testfile"))
			if tt.content != "" {
				buffer, err := os.ReadFile(overlay.File("testfile"))
				assert.NoError(t, err)
				assert.Equal(t, tt.content, string(buffer))
			}
			assert.Equal(t, tt.exists, overlay.Exists())
			assert.Equal(t, tt.isSite, overlay.IsSiteOverlay())
			assert.Equal(t, tt.isDist, overlay.IsDistributionOverlay())
		})
	}
}

func Test_BuildOverlayIndir(t *testing.T) {
	tests := map[string]struct {
		node            node.Node
		overlays        []string
		overlayFiles    map[string]string
		overlayDirs     []string
		overlaySymlinks map[string]string
		outputDir       string
		outputFiles     map[string]string
		outputDirs      []string
		outputSymlinks  map[string]string
	}{
		"empty": {
			outputDir: "/image",
		},
		"empty directory": {
			overlays:    []string{"o1"},
			overlayDirs: []string{"/var/lib/warewulf/overlays/o1/rootfs/testdir"},
			outputDir:   "/image",
			outputDirs:  []string{"testdir"},
		},
		"single flat file": {
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/testfile": "A test file from o1",
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"testfile": "A test file from o1",
			},
		},
		"multiple overlays": {
			overlays: []string{"o1", "o2"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/t1.txt": "A test file from o1",
				"/var/lib/warewulf/overlays/o2/rootfs/t2.txt": "A test file from o2",
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"t1.txt": "A test file from o1",
				"t2.txt": "A test file from o2",
			},
		},
		"template": {
			node:     node.Node{Profile: node.Profile{Comment: "A node comment"}},
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/t1.txt.ww": "{{ .Comment }}",
				"/image/t1.txt": "Previous content",
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"t1.txt":          "A node comment",
				"t1.txt.wwbackup": "Previous content",
			},
		},
		"multifile": {
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/file.txt.ww": `
{{ file "t1.txt" }}
T1
{{ file "t2.txt" }}
T2
{{ file "t3.txt" }}
T3
`,
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"t1.txt": "\nT1\n",
				"t2.txt": "T2\n",
				"t3.txt": "T3\n",
			},
		},
		"symlink": {
			overlays: []string{"o1"},
			overlaySymlinks: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/test-link": "/test-target",
			},
			outputDir: "/image",
			outputSymlinks: map[string]string{
				"test-link": "/test-target",
			},
		},
		"symlink from template": {
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/test-link.ww": `
{{- softlink "/test-target" }}
`,
			},
			outputDir: "/image",
			outputSymlinks: map[string]string{
				"test-link": "/test-target",
			},
		},
		"multiple symlinks from template": {
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/test-link.ww": `
{{ file "test-link1"}}
{{ softlink "/test-target1" }}
{{ file "test-link2"}}
{{ softlink "/test-target2" }}
`,
			},
			outputDir: "/image",
			outputSymlinks: map[string]string{
				"test-link1": "/test-target1",
				"test-link2": "/test-target2",
			},
		},
		"expansion of nodes": {
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/node.txt.ww": `
IPMI user:{{ .Ipmi.UserName}}
Kernel Version:{{.Kernel.Version}}
Kernel Args:{{.Kernel.Args | join " "}}
NetDevs:{{.NetDevs}}
Tags:{{.Tags}}
`,
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"node.txt": `
IPMI user:
Kernel Version:
Kernel Args:
NetDevs:map[]
Tags:map[]
`,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()

			for fileName, content := range tt.overlayFiles {
				env.WriteFile(fileName, content)
			}
			for _, dirName := range tt.overlayDirs {
				env.MkdirAll(dirName)
			}
			for linkName, target := range tt.overlaySymlinks {
				env.Symlink(target, linkName)
			}

			env.MkdirAll(tt.outputDir)

			assert.NoError(t, BuildOverlayIndir(tt.node, []node.Node{tt.node}, tt.overlays, env.GetPath(tt.outputDir)))
			dirFiles := tt.outputDirs
			for outputFile := range tt.outputFiles {
				dirFiles = append(dirFiles, outputFile)
			}
			for outputFile := range tt.outputSymlinks {
				dirFiles = append(dirFiles, outputFile)
			}
			sort.Strings(dirFiles)
			assert.Equal(t, dirFiles, env.ReadDir(tt.outputDir))
			for fileName, content := range tt.outputFiles {
				assert.Equal(t, content, env.ReadFile(path.Join(tt.outputDir, fileName)))
			}
			for _, dirName := range tt.outputDirs {
				assert.True(t, util.IsDir(env.GetPath(path.Join(tt.outputDir, dirName))), fmt.Sprintf("%s is not a directory", dirName))
			}
			for linkName, expTarget := range tt.outputSymlinks {
				target, err := os.Readlink(env.GetPath(path.Join(tt.outputDir, linkName)))
				assert.NoError(t, err)
				assert.Equal(t, expTarget, target)
			}
		})
	}
}

func Test_BuildOverlay(t *testing.T) {
	tests := []struct {
		description string
		nodeName    string
		context     string
		overlays    []string
		image       string
		contents    []string
		perms       []int
		hasFiles    bool
	}{
		{
			description: "if no node, context, or overlays are specified then no overlay image is generated",
			nodeName:    "",
			context:     "",
			overlays:    nil,
			image:       "",
			contents:    nil,
		},
		{
			description: "if only node is specified then no overlay image is generated",
			nodeName:    "node1",
			context:     "",
			overlays:    nil,
			image:       "",
			contents:    nil,
		},
		{
			description: "if only context is specified then context named image is generated",
			nodeName:    "",
			context:     "system",
			overlays:    nil,
			image:       "__SYSTEM__.img",
			contents:    nil,
		},
		{
			description: "if an overlay is specified without a node, then the overlay is built directly in the overlay directory",
			nodeName:    "",
			context:     "",
			overlays:    []string{"o1"},
			image:       "o1.img",
			contents:    []string{"o1.txt"},
			perms:       []int{0o644},
		},
		{
			description: "if multiple overlays are specified without a node, then the combined overlay is built directly in the overlay directory",
			nodeName:    "",
			context:     "",
			overlays:    []string{"o1", "o2"},
			image:       "o1-o2.img",
			contents:    []string{"o1.txt", "o2.txt"},
			perms:       []int{0o644, 0o644},
		},
		{
			description: "if a single node overlay is specified, then the overlay is built in a node overlay directory",
			nodeName:    "node1",
			context:     "",
			overlays:    []string{"o1"},
			image:       "node1/o1.img",
			contents:    []string{"o1.txt"},
			perms:       []int{0o644},
		},
		{
			description: "if multiple node overlays are specified, then the combined overlay is built in a node overlay directory",
			nodeName:    "node1",
			context:     "",
			overlays:    []string{"o1", "o2"},
			image:       "node1/o1-o2.img",
			contents:    []string{"o1.txt", "o2.txt"},
			perms:       []int{0o644, 0o644},
		},
		{
			description: "if no node system overlays are specified, then context pointed overlay is generated",
			nodeName:    "node1",
			context:     "system",
			overlays:    nil,
			image:       "node1/__SYSTEM__.img",
			contents:    nil,
		},
		{
			description: "if no node runtime overlays are specified, then context pointed overlay is generated",
			nodeName:    "node1",
			context:     "runtime",
			overlays:    nil,
			image:       "node1/__RUNTIME__.img",
			contents:    nil,
		},
		{
			description: "if a single node system overlay is specified, then a system overlay image is generated in a node overlay directory",
			nodeName:    "node1",
			context:     "system",
			overlays:    []string{"o1"},
			image:       "node1/__SYSTEM__.img",
			contents:    []string{"o1.txt"},
			perms:       []int{0o644},
		},
		{
			description: "if a single node runtime overlay is specified, then a runtime overlay image is generated in a node overlay directory",
			nodeName:    "node1",
			context:     "runtime",
			overlays:    []string{"o1"},
			image:       "node1/__RUNTIME__.img",
			contents:    []string{"o1.txt"},
			perms:       []int{0o644},
		},
		{
			description: "if multiple node system overlays are specified, then a system overlay image is generated with the contents of both overlays",
			nodeName:    "node1",
			context:     "system",
			overlays:    []string{"o1", "o2"},
			image:       "node1/__SYSTEM__.img",
			contents:    []string{"o1.txt", "o2.txt"},
			perms:       []int{0o644, 0o644},
		},
		{
			description: "if multiple node runtime overlays are specified, then a runtime overlay image is generated with the contents of both overlays",
			nodeName:    "node1",
			context:     "runtime",
			overlays:    []string{"o1", "o2"},
			image:       "node1/__RUNTIME__.img",
			contents:    []string{"o1.txt", "o2.txt"},
			perms:       []int{0o644, 0o644},
		},
		{
			description: "validating altered permissions are retained",
			nodeName:    "node1",
			context:     "runtime",
			overlays:    []string{"o3"},
			image:       "node1/__RUNTIME__.img",
			contents:    []string{"subdir", "subdir/o3.txt"},
			perms:       []int{0o700, 0o600},
		},
	}

	env := testenv.New(t)
	defer env.RemoveAll()

	env.CreateFile("var/lib/warewulf/overlays/o1/rootfs/o1.txt")
	env.Chmod("var/lib/warewulf/overlays/o1/rootfs/o1.txt", 0o644)
	env.CreateFile("var/lib/warewulf/overlays/o2/rootfs/o2.txt")
	env.Chmod("var/lib/warewulf/overlays/o2/rootfs/o2.txt", 0o644)
	env.CreateFile("var/lib/warewulf/overlays/o3/rootfs/subdir/o3.txt.ww")
	env.Chmod("var/lib/warewulf/overlays/o3/rootfs/subdir", 0o700)
	env.Chmod("var/lib/warewulf/overlays/o3/rootfs/subdir/o3.txt.ww", 0o600)

	for _, tt := range tests {
		nodeInfo := node.NewNode(tt.nodeName)
		t.Run(tt.description, func(t *testing.T) {
			err := BuildOverlay(nodeInfo, []node.Node{nodeInfo}, tt.context, tt.overlays)
			assert.NoError(t, err)
			if tt.image != "" {
				image := env.GetPath(path.Join("srv/warewulf/overlays", tt.image))
				assert.FileExists(t, image)
				sort.Strings(tt.contents)
				headers, err := readCpio(image)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.contents), len(headers))
				for i, file := range tt.contents {
					assert.Equal(t, file, headers[file].Name)
					if len(tt.perms) > i {
						assert.Equal(t, tt.perms[i], int(headers[file].Mode.Perm()))
					}
				}
			} else {
				dirName := env.GetPath(path.Join("srv/warewulf/overlays", tt.nodeName))
				isEmpty := dirIsEmpty(t, dirName)
				assert.True(t, isEmpty, "%v should be empty, but isn't", dirName)
			}
		})
	}
}

func Test_BuildAllOverlays(t *testing.T) {
	tests := []struct {
		description     string
		nodes           []string
		systemOverlays  [][]string
		runtimeOverlays [][]string
		createdOverlays []string
	}{
		{
			description:     "empty input creates no overlays",
			nodes:           nil,
			systemOverlays:  nil,
			runtimeOverlays: nil,
			createdOverlays: nil,
		},
		{
			description:     "a node with no overlays creates default system/runtime overlays",
			nodes:           []string{"node1"},
			systemOverlays:  [][]string{{"o1"}},
			runtimeOverlays: [][]string{{"o1"}},
			createdOverlays: []string{"node1/__SYSTEM__.img.gz", "node1/__RUNTIME__.img.gz"},
		},
		{
			description:     "multiple nodes with no overlays creates default system/runtime overlays",
			nodes:           []string{"node1", "node2"},
			systemOverlays:  [][]string{{"o1"}, {"o2"}},
			runtimeOverlays: [][]string{{"o1"}, {"o2"}},
			createdOverlays: []string{"node1/__SYSTEM__.img.gz", "node1/__RUNTIME__.img.gz", "node2/__SYSTEM__.img.gz", "node2/__RUNTIME__.img.gz"},
		},
		{
			description:     "a system overlay for a node generates a system overlay for that node",
			nodes:           []string{"node1"},
			systemOverlays:  [][]string{{"o1"}},
			runtimeOverlays: nil,
			createdOverlays: []string{"node1/__SYSTEM__.img.gz"},
		},
		{
			description:     "two nodes with different system overlays generates a system overlay for each node",
			nodes:           []string{"node1", "node2"},
			systemOverlays:  [][]string{{"o1"}, {"o1", "o2"}},
			runtimeOverlays: nil,
			createdOverlays: []string{"node1/__SYSTEM__.img.gz", "node2/__SYSTEM__.img.gz"},
		},
		{
			description:     "two nodes with a single runtime overlay generates a runtime overlay for the first node",
			nodes:           []string{"node1"},
			systemOverlays:  nil,
			runtimeOverlays: [][]string{{"o1"}},
			createdOverlays: []string{"node1/__RUNTIME__.img.gz"},
		},
		{
			description:     "two nodes with different runtime overlays generates a system overlay for each node",
			nodes:           []string{"node1", "node2"},
			systemOverlays:  nil,
			runtimeOverlays: [][]string{{"o1"}, {"o1", "o2"}},
			createdOverlays: []string{"node1/__RUNTIME__.img.gz", "node2/__RUNTIME__.img.gz"},
		},
		{
			description:     "a node with both a runtime and system overlay generates an image for each",
			nodes:           []string{"node1"},
			systemOverlays:  [][]string{{"o1"}},
			runtimeOverlays: [][]string{{"o2"}},
			createdOverlays: []string{"node1/__RUNTIME__.img.gz", "node1/__SYSTEM__.img.gz"},
		},
		{
			description:     "two nodes with both runtime and system overlays generates each image for each node",
			nodes:           []string{"node1", "node2"},
			systemOverlays:  [][]string{{"o1"}, {"o1", "o2"}},
			runtimeOverlays: [][]string{{"o2"}, {"o2"}},
			createdOverlays: []string{"node1/__RUNTIME__.img.gz", "node1/__SYSTEM__.img.gz", "node2/__RUNTIME__.img.gz", "node2/__SYSTEM__.img.gz"},
		},
	}

	conf := warewulfconf.Get()
	overlayDir, overlayDirErr := os.MkdirTemp(os.TempDir(), "ww-test-overlay-*")
	assert.NoError(t, overlayDirErr)
	defer os.RemoveAll(overlayDir)
	conf.Paths.WWOverlaydir = overlayDir
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o1"), 0o700))
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o2"), 0o700))

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			provisionDir, provisionDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
			assert.NoError(t, provisionDirErr)
			defer os.RemoveAll(provisionDir)
			conf.Paths.WWProvisiondir = provisionDir

			var nodes []node.Node
			for i, nodeName := range tt.nodes {
				nodeInfo := node.NewNode(nodeName)
				if tt.systemOverlays != nil {
					nodeInfo.SystemOverlay = tt.systemOverlays[i]
				}
				if tt.runtimeOverlays != nil {
					nodeInfo.RuntimeOverlay = tt.runtimeOverlays[i]
				}
				nodes = append(nodes, nodeInfo)
			}
			err := BuildAllOverlays(nodes, nodes, runtime.NumCPU())
			assert.NoError(t, err)
			if tt.createdOverlays == nil {
				dirName := path.Join(provisionDir, "overlays")
				assert.True(t, dirIsEmpty(t, dirName), "%v should be empty, but isn't", dirName)
			}
			for _, overlayPath := range tt.createdOverlays {
				assert.FileExists(t, path.Join(provisionDir, "overlays", overlayPath))
			}
		})
	}
}

func Test_BuildSpecificOverlays(t *testing.T) {
	tests := []struct {
		description string
		nodes       []string
		overlays    []string
		images      []string
		succeed     bool
	}{
		{
			description: "building no overlays for no nodes generates no error and no images",
			nodes:       nil,
			overlays:    nil,
			images:      nil,
			succeed:     true,
		},
		{
			description: "building no overlays for a node generates no error and no images",
			nodes:       []string{"node1"},
			overlays:    nil,
			images:      nil,
			succeed:     true,
		},
		{
			description: "building no overlays for two nodes generates no error and no images",
			nodes:       []string{"node1", "node2"},
			overlays:    nil,
			images:      nil,
			succeed:     true,
		},
		{
			description: "building an overlay for a node generates an overlay image in that node's overlay directory",
			nodes:       []string{"node1"},
			overlays:    []string{"o1"},
			images:      []string{"node1/o1.img"},
			succeed:     true,
		},
		{
			description: "building an overlay for two nodes generates an overlay image in each node's overlay directory",
			nodes:       []string{"node1", "node2"},
			overlays:    []string{"o1"},
			images:      []string{"node1/o1.img", "node2/o1.img"},
			succeed:     true,
		},
		{
			description: "building multiple overlays for a node generates an overlay image for each overlay in the node's overlay directory",
			nodes:       []string{"node1"},
			overlays:    []string{"o1", "o2"},
			images:      []string{"node1/o1.img", "node1/o2.img"},
			succeed:     true,
		},
		{
			description: "building multiple overlays for two nodes generates an overlay image for each overlay in each node's overlay directory",
			nodes:       []string{"node1", "node2"},
			overlays:    []string{"o1", "o2"},
			images:      []string{"node1/o1.img", "node1/o2.img", "node2/o1.img", "node2/o2.img"},
			succeed:     true,
		},
	}

	conf := warewulfconf.Get()
	overlayDir, overlayDirErr := os.MkdirTemp(os.TempDir(), "ww-test-overlay-*")
	assert.NoError(t, overlayDirErr)
	defer os.RemoveAll(overlayDir)
	conf.Paths.WWOverlaydir = overlayDir
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o1"), 0o700))
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o2"), 0o700))

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			provisionDir, provisionDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
			assert.NoError(t, provisionDirErr)
			defer os.RemoveAll(provisionDir)
			conf.Paths.WWProvisiondir = provisionDir

			var nodes []node.Node
			for _, nodeName := range tt.nodes {
				nodeInfo := node.NewNode(nodeName)
				nodes = append(nodes, nodeInfo)
			}
			err := BuildSpecificOverlays(nodes, nodes, tt.overlays, runtime.NumCPU())
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

func Test_CreateOverlayFile(t *testing.T) {
	tests := []struct {
		name        string
		overlayName string
		filePath    string
		content     []byte
		force       bool
	}{
		{"create file", "test", "newfile.ww", []byte("new file"), false},
		{"overwrite existing file", "test", "existingfile.ww", []byte("overwrite file"), true},
	}

	conf := warewulfconf.Get()
	overlayDir, overlayDirErr := os.MkdirTemp(os.TempDir(), "ww-test-overlay-*")
	assert.NoError(t, overlayDirErr)
	defer os.RemoveAll(overlayDir)
	conf.Paths.WWOverlaydir = overlayDir

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newOverlay := GetSiteOverlay(tt.overlayName)
			err := newOverlay.AddFile(tt.filePath, tt.content, true, tt.force)
			assert.NoError(t, err)

			newFile := newOverlay.File(tt.filePath)
			assert.FileExists(t, newFile)
			readContent, err := os.ReadFile(newFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.content, readContent)
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

func readCpio(name string) (headers map[string]*cpio.Header, err error) {
	f, err := os.Open(name)
	if err != nil {
		return headers, err
	}
	defer f.Close()

	reader := cpio.NewReader(f)
	headers = make(map[string]*cpio.Header)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return headers, nil
		}
		if err != nil {
			return headers, err
		}
		headers[header.Name] = header
	}
}
