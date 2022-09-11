package overlay

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
Reads a file file from the host fs. If the file has nor '/' prefix
the path is relative to SYSCONFDIR.
Templates in the file are no evaluated.
*/
func templateFileInclude(inc string) string {
	if !strings.HasPrefix(inc, "/") {
		inc = path.Join(buildconfig.SYSCONFDIR(), "warewulf", inc)
	}
	wwlog.Debug("Including file into template: %s\n", inc)
	content, err := ioutil.ReadFile(inc)
	if err != nil {
		wwlog.Verbose("Could not include file into template: %s\n", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}

/*
Reads a file into template the abort string is found in a line. First argument
is the file to read, the second the abort string
Templates in the file are no evaluated.
*/
func templateFileBlock(inc string, abortStr string) (string, error) {
	if !strings.HasPrefix(inc, "/") {
		inc = path.Join(buildconfig.SYSCONFDIR(), "warewulf", inc)
	}
	wwlog.Debug("Including file block into template: %s\n", inc)
	readFile, err := os.Open(inc)
	if err != nil {
		return "", err
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

/*
Reads a file relative to given container.
Templates in the file are no evaluated.
*/
func templateContainerFileInclude(containername string, filepath string) string {
	wwlog.Verbose("Including file from Container into template: %s:%s\n", containername, filepath)

	if containername == "" {
		wwlog.Warn("Container is not defined for node: %s\n", filepath)
		return ""
	}

	if !container.ValidSource(containername) {
		wwlog.Warn("Template requires file(s) from non-existant container: %s:%s\n", containername, filepath)
		return ""
	}

	containerDir := container.RootFsDir(containername)

	wwlog.Debug("Including file from container: %s:%s\n", containerDir, filepath)

	if !util.IsFile(path.Join(containerDir, filepath)) {
		wwlog.Warn("Requested file from container does not exist: %s:%s\n", containername, filepath)
		return ""
	}

	content, err := ioutil.ReadFile(path.Join(containerDir, filepath))

	if err != nil {
		wwlog.Error("Template include failed: %s\n", err)
	}
	return strings.TrimSuffix(string(content), "\n")
}
