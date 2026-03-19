package warewulfd

import (
	"net/http"
	"os"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// noListFileSystem wraps http.FileSystem and returns os.ErrNotExist for
// directories, disabling directory listing.
type noListFileSystem struct {
	http.FileSystem
}

func (fs noListFileSystem) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	if stat.IsDir() {
		_ = f.Close()
		return nil, os.ErrNotExist
	}

	return f, nil
}

// HandleFiles serves static files from the configured warewulf files directory.
// Subdirectories are supported. Directory listing is disabled.
func HandleFiles(w http.ResponseWriter, req *http.Request) {
	conf := warewulfconf.Get()
	filesDir := conf.Paths.WWFilesdir
	wwlog.Debug("Serving file from %s: %s", filesDir, req.URL.Path)
	fs := noListFileSystem{http.Dir(filesDir)}
	http.StripPrefix("/files/", http.FileServer(fs)).ServeHTTP(w, req)
}
