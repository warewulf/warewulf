package overlay

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// Reads a file file from the host fs. If the file has nor '/' prefix the path
// is relative to Paths.Sysconfdir. Templates in the file are no evaluated.
func templateFileInclude(inc string) string {
	conf := warewulfconf.Get()
	if !strings.HasPrefix(inc, "/") {
		inc = path.Join(conf.Paths.Sysconfdir, "warewulf", inc)
	}
	wwlog.Debug("Including file into template: %s", inc)
	content, err := os.ReadFile(inc)
	if err != nil {
		wwlog.Verbose("Could not include file into template: %s", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}

// Reads a file into template the abort string is found in a line. First
// argument is the file to read, the second the abort string. Templates in the
// file are no evaluated.
func templateFileBlock(inc string, abortStr string) (string, error) {
	conf := warewulfconf.Get()
	if !strings.HasPrefix(inc, "/") {
		inc = path.Join(conf.Paths.Sysconfdir, "warewulf", inc)
	}
	wwlog.Debug("Including file block into template: %s", inc)
	readFile, err := os.Open(inc)
	if err != nil {
		wwlog.Info("couldn't read block %s: %s", inc, err)
		return abortStr, nil
	}
	defer readFile.Close()
	var cont string
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if strings.Contains(line, abortStr) {
			break
		}
		cont += line + "\n"
	}

	// NOTE: the text originally contains N-1 newlines for N lines, but the above
	// loop will always add one at the end
	// Avoids adding a blank line that was not present in the original file
	// by adding 'abort' string to the end of the included block (without a newline)
	// instead of manually in the template
	cont += abortStr

	return cont, nil

}

// Reads a file relative to given image. Templates in the file are not evaluated.
func templateImageFileInclude(imagename string, filepath string) string {
	wwlog.Verbose("Including file from Image into template: %s:%s", imagename, filepath)

	if imagename == "" {
		wwlog.Warn("Image is not defined for node: %s", filepath)
		return ""
	}

	if !image.ValidSource(imagename) {
		wwlog.Warn("Template requires file(s) from non-existant image: %s:%s", imagename, filepath)
		return ""
	}

	imageDir := image.RootFsDir(imagename)

	wwlog.Debug("Including file from image: %s:%s", imageDir, filepath)

	if !util.IsFile(path.Join(imageDir, filepath)) {
		wwlog.Warn("Requested file from image does not exist: %s:%s", imagename, filepath)
		return ""
	}

	content, err := os.ReadFile(path.Join(imageDir, filepath))

	if err != nil {
		wwlog.Error("Template include failed: %s", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}

// Don't return an error as we use this function for template evaluation, so
// error will turn up there as the return string
func createIgnitionJson(node *node.Node) string {
	conf, rep, err := node.GetConfig()
	if len(conf.Storage.Disks) == 0 && len(conf.Storage.Filesystems) == 0 {
		wwlog.Debug("no disks or filesystems present, don't create a json object")
		return ""
	}
	if err != nil {
		wwlog.Error("disk, filesystem configuration has following error: ", fmt.Sprint(err))
		return fmt.Sprint(err)
	}
	if rep != "" {
		wwlog.Warn("%s storage configuration has following non fatal problems: %s", node.Id, rep)
	}
	tmpYaml, _ := json.Marshal(&conf)
	return string(tmpYaml)
}

func importSoftlink(lnk string) string {
	target, err := filepath.EvalSymlinks(lnk)
	if err != nil {
		return "abort"
	}
	wwlog.Debug("importing softlink pointing to: %s", target)
	return softlink(target)
}

func softlink(target string) string {
	return fmt.Sprintf("{{ /* softlink \"%s\" */ }}", target)
}

// UniqueField returns a filtered version of a multi-line input string. input is
// expected to be a field-separated format with one record per line (terminated
// by `\n`). Order of lines is preserved, with the first matching line taking
// precedence.
//
// For example, parsing /etc/passwd filter /etc/passwd for unique user names:
//
// Lines without the index field (e.g., blank lines) are always included in the
// output.
//
// UniqueField(":", 0, passwdContent)
func UniqueField(sep string, index int, input string) string {
	inputLines := strings.Split(input, "\n")
	var outputLines []string
	found := make(map[string]bool)
	for _, line := range inputLines {
		inputFields := strings.Split(line, sep)
		if len(inputFields) > index {
			field := inputFields[index]
			if field != "" {
				if found[field] {
					continue
				}
				found[field] = true
			}
		}
		outputLines = append(outputLines, line)
	}
	return strings.Join(outputLines, "\n")
}
