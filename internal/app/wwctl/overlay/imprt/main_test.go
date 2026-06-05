package imprt

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_List(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "warewulf")
	if err != nil {
		t.Errorf("Could not create temp folder: %v", err)
		t.FailNow()
	}
	defer os.RemoveAll(tmpdir)

	overlayDir := fmt.Sprintf("%s/overlay", tmpdir)
	err = os.MkdirAll(overlayDir, 0o755)
	if err != nil {
		t.Errorf("Could not create target folder: %s, err: %v", overlayDir, err)
		t.FailNow()
	}

	importDir := fmt.Sprintf("%s/test", overlayDir)
	err = os.MkdirAll(importDir, 0o755)
	if err != nil {
		t.Errorf("Could not create target folder: %s, err: %v", importDir, err)
		t.FailNow()
	}

	file, err := os.CreateTemp(tmpdir, "file")
	if err != nil {
		t.Errorf("Could not create tempfile")
		t.FailNow()
	}
	file.Close()
	err = os.Chmod(file.Name(), 0o755)
	if err != nil {
		t.Errorf("Could not change the file %s mode: %v", file.Name(), err)
		t.FailNow()
	}

	inDb := `nodeprofiles:
  default: {}
nodes: {}
`
	conf_yml := ``

	conf := warewulfconf.New()
	err = conf.Parse([]byte(conf_yml), false)
	assert.NoError(t, err)
	warewulfd.SetNoDaemon()
	conf.Paths.WWOverlaydir = overlayDir

	_, err = node.Parse([]byte(inDb))
	assert.NoError(t, err)
	t.Logf("Running test: wwctl overlay import test\n")
	t.Run("wwctl overlay import test", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-n", "test", file.Name()})
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		err = baseCmd.Execute()
		if err == nil {
			t.Errorf("Should recieve error when running command")
			t.FailNow()
		}
		if _, err = os.Stat(importDir + file.Name()); err == nil {
			t.Errorf("Target file %s should not exist", importDir+file.Name())
			t.FailNow()
		}

		baseCmd.SetArgs([]string{"-p", "test", file.Name()})
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		err = baseCmd.Execute()
		if err != nil {
			t.Errorf("Received error when running command, err: %v\n", err)
			t.FailNow()
		}
		if _, err = os.Stat(importDir + file.Name()); os.IsNotExist(err) {
			t.Errorf("Target file %s should exist", importDir+file.Name())
			t.FailNow()
		}
	})
}

func Test_Import(t *testing.T) {
	tests := map[string]struct {
		initFiles   []string
		initDirs    []string
		args        []string
		errExpected bool
	}{
		"import a file": {
			initFiles: []string{"importfile"},
			initDirs:  []string{"/var/lib/warewulf/overlays/to1/rootfs"},
			args:      []string{"to1", "importfile"},
		},

		"import missing parent": {
			initFiles:   []string{"importfile"},
			initDirs:    []string{"/var/lib/warewulf/overlays/to1/rootfs"},
			args:        []string{"to1", "importfile", "a/b/importfile"},
			errExpected: true,
		},

		"import create parents": {
			initFiles:   []string{"importfile"},
			initDirs:    []string{"/var/lib/warewulf/overlays/to1/rootfs"},
			args:        []string{"to1", "importfile", "a/b/importfile", "--parents"},
			errExpected: false,
		},

		"import fail overwrite": {
			initFiles:   []string{"importfile", "/var/lib/warewulf/overlays/to1/rootfs/importfile"},
			args:        []string{"to1", "importfile"},
			errExpected: true,
		},

		"import overwrite": {
			initFiles:   []string{"importfile", "/var/lib/warewulf/overlays/to1/rootfs/importfile"},
			args:        []string{"to1", "importfile", "--overwrite"},
			errExpected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			{
				wd, err := os.Getwd()
				assert.NoError(t, err)
				defer func() { assert.NoError(t, os.Chdir(wd)) }()
			}
			assert.NoError(t, os.Chdir(env.GetPath(".")))

			OverwriteFile = false
			CreateDirs = false
			ArchiveImport = false

			for _, file := range tt.initFiles {
				env.CreateFile(file)
			}
			for _, dir := range tt.initDirs {
				env.MkdirAll(dir)
			}

			cmd := GetCommand()
			cmd.SetArgs(tt.args)
			stdout := new(bytes.Buffer)
			cmd.SetOut(stdout)
			stderr := new(bytes.Buffer)
			cmd.SetErr(stderr)
			err := cmd.Execute()
			if tt.errExpected {
				assert.Error(t, err, stdout)
			} else {
				assert.NoError(t, err, stderr)
			}
		})
	}
}

func Test_ImportArchive(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	overlayName := "archive-overlay"
	archivePath := env.GetPath("artifact.tar.gz")
	createOverlayArchive(t, archivePath, overlayName, map[string]string{
		"rootfs/etc/config": "hello\n",
	})

	OverwriteFile = false
	CreateDirs = false
	ArchiveImport = false

	cmd := GetCommand()
	cmd.SetArgs([]string{"--archive", overlayName, archivePath})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, cmd.Execute()) {
		return
	}

	data, err := os.ReadFile(env.GetPath(filepath.Join("var/lib/warewulf/overlays", overlayName, "rootfs/etc/config")))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "hello\n", string(data))
	assert.FileExists(t, env.GetPath(filepath.Join("var/lib/warewulf/overlays", overlayName, overlaydiff.ArtifactManifestFileName)))
	assert.Contains(t, overlay.FindOverlays(), overlayName)
}

func Test_ImportArchiveOverwrite(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	overlayName := "archive-overlay"
	archivePath := env.GetPath("artifact.tar.gz")
	createOverlayArchive(t, archivePath, overlayName, map[string]string{
		"rootfs/etc/config": "new\n",
	})
	env.WriteFile(filepath.Join("var/lib/warewulf/overlays", overlayName, "rootfs/etc/config"), "old\n")

	OverwriteFile = false
	CreateDirs = false
	ArchiveImport = false

	cmd := GetCommand()
	cmd.SetArgs([]string{"--archive", overlayName, archivePath})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	err := cmd.Execute()
	assert.Error(t, err)

	OverwriteFile = false
	CreateDirs = false
	ArchiveImport = false
	cmd = GetCommand()
	cmd.SetArgs([]string{"--archive", "--overwrite", overlayName, archivePath})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, cmd.Execute()) {
		return
	}
	assert.Equal(t, "new\n", env.ReadFile(filepath.Join("var/lib/warewulf/overlays", overlayName, "rootfs/etc/config")))
}

func Test_ImportArchiveRejectsNameMismatch(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	archivePath := env.GetPath("artifact.tar.gz")
	createOverlayArchive(t, archivePath, "manifest-name", map[string]string{
		"rootfs/etc/config": "hello\n",
	})

	OverwriteFile = false
	CreateDirs = false
	ArchiveImport = false

	cmd := GetCommand()
	cmd.SetArgs([]string{"--archive", "requested-name", archivePath})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match")
}

func Test_ImportArchiveRejectsTraversal(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	overlayName := "archive-overlay"
	archivePath := env.GetPath("artifact.tar.gz")
	createOverlayArchive(t, archivePath, overlayName, map[string]string{
		"../escape":       "bad",
		"rootfs/etc/file": "ok",
	})

	OverwriteFile = false
	CreateDirs = false
	ArchiveImport = false

	cmd := GetCommand()
	cmd.SetArgs([]string{"--archive", overlayName, archivePath})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	err := cmd.Execute()
	assert.Error(t, err)
	assert.NoDirExists(t, env.GetPath(filepath.Join("var/lib/warewulf/overlays", overlayName)))
}

func createOverlayArchive(t *testing.T, archivePath string, overlayName string, files map[string]string) {
	t.Helper()
	archiveFile, err := os.Create(archivePath)
	assert.NoError(t, err)
	defer archiveFile.Close()

	gzipWriter := gzip.NewWriter(archiveFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	writeTarDir(t, tarWriter, "rootfs")
	manifest := overlaydiff.BuildArtifactManifest(overlayName, "/", "", []string{"/etc/config"}, overlaydiff.DecisionSummary{Selected: 1})
	manifestFile, err := os.CreateTemp(t.TempDir(), "manifest-*.json")
	assert.NoError(t, err)
	manifestPath := manifestFile.Name()
	assert.NoError(t, manifestFile.Close())
	assert.NoError(t, overlaydiff.SaveArtifactManifest(manifestPath, manifest))
	writeTarFileFromDisk(t, tarWriter, overlaydiff.ArtifactManifestFileName, manifestPath)

	for name, content := range files {
		writeTarFile(t, tarWriter, name, content)
	}
}

func writeTarDir(t *testing.T, writer *tar.Writer, name string) {
	t.Helper()
	assert.NoError(t, writer.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeDir, Mode: 0o755}))
}

func writeTarFile(t *testing.T, writer *tar.Writer, name string, content string) {
	t.Helper()
	header := &tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(content))}
	assert.NoError(t, writer.WriteHeader(header))
	_, err := io.WriteString(writer, content)
	assert.NoError(t, err)
}

func writeTarFileFromDisk(t *testing.T, writer *tar.Writer, name string, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	writeTarFile(t, writer, name, string(data))
}
