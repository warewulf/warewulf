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
			overlay, err := Get(tt.name)
			if tt.exists {
				assert.NoError(t, err)
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
			}
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
				"t2.txt": "\nT2\n",
				"t3.txt": "\nT3\n",
			},
		},
		"multifile whitespace trimmed": {
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/file.txt.ww": `{{- range $i, $name := list "a" "b" "c" }}
{{ file $name }}
{{- end -}}`,
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"a": "\n",
				"b": "\n",
				"c": "",
			},
		},
		"abort": {
			// Regression test: abort() must suppress all file output in BuildOverlayIndir.
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/file.txt.ww": `{{- abort -}}`,
			},
			outputDir:   "/image",
			outputFiles: map[string]string{},
		},
		"multifile all empty with whitespace trimming": {
			// Regression test: before the state-based rewrite, using {{- file "name" -}}
			// caused adjacent file() sentinels to collapse onto one line. The greedy .*
			// in the regex matched only the last sentinel, so only the final file was
			// created. All three files must be created here, each with zero content.
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/file.txt.ww": `{{- range $name := list "a" "b" "c" -}}
{{- file $name -}}
{{- end -}}`,
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"a": "",
				"b": "",
				"c": "",
			},
		},
		"multifile default symlink written to disk": {
			// A softlink() call before any file() call targets the default output
			// path and must be created even when named files are also present.
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/file.txt.ww": `{{- softlink "/link-target" -}}{{- file "named.txt" -}}named content`,
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"named.txt": "named content",
			},
			outputSymlinks: map[string]string{
				"file.txt": "/link-target",
			},
		},
		"multifile pre-file content not written to disk": {
			// Content written before the first file() call is preserved in the
			// RenderedTemplate (RenderTemplateFile returns it), but BuildOverlayIndir
			// does not write it to disk when named files are present. Only named.txt
			// is created; file.txt is not, despite the non-empty default buffer.
			overlays: []string{"o1"},
			overlayFiles: map[string]string{
				"/var/lib/warewulf/overlays/o1/rootfs/file.txt.ww": `pre-file content
{{ file "named.txt" }}named content`,
			},
			outputDir: "/image",
			outputFiles: map[string]string{
				"named.txt": "named content",
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
	defer func() { _ = os.RemoveAll(overlayDir) }()
	conf.Paths.WWOverlaydir = overlayDir
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o1"), 0o700))
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o2"), 0o700))

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			provisionDir, provisionDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
			assert.NoError(t, provisionDirErr)
			defer func() { _ = os.RemoveAll(provisionDir) }()
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
	defer func() { _ = os.RemoveAll(overlayDir) }()
	conf.Paths.WWOverlaydir = overlayDir
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o1"), 0o700))
	assert.NoError(t, os.Mkdir(path.Join(overlayDir, "o2"), 0o700))

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			provisionDir, provisionDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
			assert.NoError(t, provisionDirErr)
			defer func() { _ = os.RemoveAll(provisionDir) }()
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
		{"create file", "test", "test.ww", []byte("new file"), false},
		{"overwrite existing file", "test", "test.ww", []byte("overwrite file"), true},
	}

	conf := warewulfconf.Get()
	overlayDir, overlayDirErr := os.MkdirTemp(os.TempDir(), "ww-test-overlay-*")
	assert.NoError(t, overlayDirErr)
	defer func() { _ = os.RemoveAll(overlayDir) }()
	conf.Paths.WWOverlaydir = overlayDir
	conf.Paths.Datadir = "/dev/null"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newOverlay, err := Get(tt.overlayName)
			if err != nil {
				newOverlay, err = Create(tt.overlayName)
			}
			assert.NoError(t, err)
			err = newOverlay.AddFile(tt.filePath, tt.content, true, tt.force)
			assert.NoError(t, err)

			newFile := newOverlay.File(tt.filePath)
			assert.FileExists(t, newFile)
			readContent, err := os.ReadFile(newFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.content, readContent)
		})
	}
}

func Test_RenderTemplate_ImportLink(t *testing.T) {
	// Create a real symlink on the host filesystem for the success case.
	// ImportLink calls filepath.EvalSymlinks on the host, not in the overlay rootfs.
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "real-target")
	err := os.WriteFile(target, []byte("content"), 0644)
	assert.NoError(t, err)
	link := filepath.Join(tmpDir, "test-link")
	err = os.Symlink(target, link)
	assert.NoError(t, err)

	t.Run("ImportLink success", func(t *testing.T) {
		env := testenv.New(t)
		defer env.RemoveAll()
		tmplContent := fmt.Sprintf(`{{- ImportLink %q -}}`, link)
		tmplPath := env.GetPath("test.ww")
		env.WriteFile("test.ww", tmplContent)
		rendered, err := RenderTemplate(tmplPath, TemplateStruct{})
		assert.NoError(t, err)
		assert.True(t, rendered.Files[0].IsSymlink)
		assert.Equal(t, target, rendered.Files[0].Target)
	})

	t.Run("ImportLink failure", func(t *testing.T) {
		env := testenv.New(t)
		defer env.RemoveAll()
		tmplPath := env.GetPath("test.ww")
		env.WriteFile("test.ww", `{{- ImportLink "/nonexistent/path/that/does/not/exist" -}}`)
		_, err := RenderTemplate(tmplPath, TemplateStruct{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ImportLink")
	})
}

func TestRenderTemplate(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// Setup for Include and IncludeBlock
	includeDir := env.GetPath(path.Join(testenv.Sysconfdir, "warewulf"))
	env.MkdirAll(path.Join(testenv.Sysconfdir, "warewulf"))

	includeFile := path.Join(includeDir, "test-include")
	err := os.WriteFile(includeFile, []byte("included content\n"), 0644)
	assert.NoError(t, err)

	blockFile := path.Join(includeDir, "test-block")
	err = os.WriteFile(blockFile, []byte("line1\nline2\nABORT\nline3\n"), 0644)
	assert.NoError(t, err)

	// Setup for IncludeFrom
	imageName := "test-image"
	imageRootfs := env.GetPath(path.Join(testenv.WWChrootdir, imageName, "rootfs"))
	env.MkdirAll(path.Join(testenv.WWChrootdir, imageName, "rootfs"))
	err = os.WriteFile(path.Join(imageRootfs, "file-in-image"), []byte("image content\n"), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		template string
		data     TemplateStruct
		validate func(*testing.T, *RenderedTemplate, error)
	}{
		{
			name:     "Basic functions",
			template: `{{inc 1}} {{dec 5}} {{basename "/path/to/file"}}`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "2 4 file", rt.Files[0].Buffer.String())
			},
		},
		{
			name:     "File redirection",
			template: `default content{{file "other"}}other content`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				// Pre-file() content is non-empty, so the default slot survives the strip.
				assert.Len(t, rt.Files, 2)
				assert.Equal(t, "", rt.Files[0].Name)
				assert.Equal(t, "default content", rt.Files[0].Buffer.String())
				assert.Equal(t, "other", rt.Files[1].Name)
				assert.Equal(t, "other content", rt.Files[1].Buffer.String())
			},
		},
		{
			name:     "Softlink",
			template: `{{softlink "/target"}}`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.True(t, rt.Files[0].IsSymlink)
				assert.Equal(t, "/target", rt.Files[0].Target)
			},
		},
		{
			name:     "No backup",
			template: `{{nobackup}}content`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.False(t, rt.BackupFile)
			},
		},
		{
			name:     "Abort",
			template: `some content{{abort}}more content`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.False(t, rt.WriteFile)
				assert.Contains(t, rt.Files[0].Buffer.String(), "some content")
				assert.NotContains(t, rt.Files[0].Buffer.String(), "more content")
			},
		},
		{
			name:     "Systemd Escape",
			template: `{{SystemdEscape "foo-bar/baz"}}`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "foo\\x2dbar-baz", rt.Files[0].Buffer.String())
			},
		},
		{
			name:     "Include",
			template: `{{Include "test-include"}}`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "included content", rt.Files[0].Buffer.String())
			},
		},
		{
			name:     "IncludeBlock",
			template: `{{IncludeBlock "test-block" "ABORT"}}`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "line1\nline2\nABORT", rt.Files[0].Buffer.String())
			},
		},
		{
			name:     "IncludeFrom",
			template: fmt.Sprintf(`{{IncludeFrom %q "file-in-image"}}`, imageName),
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "image content", rt.Files[0].Buffer.String())
			},
		},
		{
			name:     "Sprig functions",
			template: `{{ "hello" | upper }}`,
			validate: func(t *testing.T, rt *RenderedTemplate, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "HELLO", rt.Files[0].Buffer.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := path.Join(env.BaseDir, tt.name+".ww")
			err := os.WriteFile(tmpFile, []byte(tt.template), 0644)
			assert.NoError(t, err)

			rt, err := RenderTemplate(tmpFile, tt.data)
			tt.validate(t, rt, err)
		})
	}
}

func TestRenderTemplateFile(t *testing.T) {
	t.Run("no file() calls returns all content", func(t *testing.T) {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.WriteFile("test.ww", `hello world`)
		buf, backupFile, writeFile, err := RenderTemplateFile(env.GetPath("test.ww"), TemplateStruct{})
		assert.NoError(t, err)
		assert.True(t, *backupFile)
		assert.True(t, *writeFile)
		assert.Equal(t, "hello world", buf.String())
	})

	t.Run("pre-file() content is returned when file() calls are present", func(t *testing.T) {
		// When a template writes content before the first file() call,
		// RenderTemplateFile returns that pre-file() content from the default slot.
		env := testenv.New(t)
		defer env.RemoveAll()
		env.WriteFile("test.ww", `pre-file content{{ file "other" }}other content`)
		buf, _, _, err := RenderTemplateFile(env.GetPath("test.ww"), TemplateStruct{})
		assert.NoError(t, err)
		assert.Equal(t, "pre-file content", buf.String())
	})

	t.Run("empty default slot is stripped when file() calls are present", func(t *testing.T) {
		// When there is no content before the first file() call, the empty
		// default slot is stripped. RenderTemplate returns only the named file.
		env := testenv.New(t)
		defer env.RemoveAll()
		env.WriteFile("test.ww", `{{- file "other" -}}other content`)
		rendered, err := RenderTemplate(env.GetPath("test.ww"), TemplateStruct{})
		assert.NoError(t, err)
		assert.Len(t, rendered.Files, 1)
		assert.Equal(t, "other", rendered.Files[0].Name)
		assert.Equal(t, "other content", rendered.Files[0].Buffer.String())
	})
}

func dirIsEmpty(t *testing.T, name string) bool {
	f, err := os.Open(name)
	if err != nil {
		t.Log(err)
		return true
	}
	defer func() { _ = f.Close() }()

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
	defer func() { _ = f.Close() }()

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
