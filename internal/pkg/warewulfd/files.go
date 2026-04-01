package warewulfd

import (
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
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

// resolveFilesPath safely resolves a request path to a file within filesDir.
// Returns the clean absolute path or an empty string if the path escapes filesDir.
func resolveFilesPath(filesDir string, reqPath string) string {
	// Strip the /files/ prefix
	relPath := strings.TrimPrefix(reqPath, "/files/")
	if relPath == "" {
		return ""
	}

	fullPath := filepath.Join(filesDir, relPath)
	cleanPath := filepath.Clean(fullPath)
	cleanDir := filepath.Clean(filesDir)

	rel, err := filepath.Rel(cleanDir, cleanPath)
	if err != nil {
		return ""
	}
	if strings.HasPrefix(rel, "..") {
		return ""
	}

	return cleanPath
}

// authenticateNode identifies and authenticates a node for the request.
// The node is identified via ?wwid= query parameter or ARP cache fallback.
// When secure mode is enabled, requests must come from a privileged port.
// If the node has an asset key, ?assetkey= must match.
// On success, returns the authenticated node and true.
// On failure, writes the HTTP error response and returns false.
func authenticateNode(w http.ResponseWriter, req *http.Request) (node.Node, bool) {
	conf := warewulfconf.Get()

	remoteAddrPort, err := netip.ParseAddrPort(req.RemoteAddr)
	if err != nil {
		wwlog.Error("could not parse remote address: %s", req.RemoteAddr)
		http.Error(w, "could not parse remote address", http.StatusInternalServerError)
		return node.Node{}, false
	}
	ipaddr := remoteAddrPort.Addr().String()
	remoteport := int(remoteAddrPort.Port())

	var hwaddr string
	if len(req.URL.Query()["wwid"]) > 0 {
		hwaddr = parseHwaddr(req.URL.Query()["wwid"][0])
	}
	if hwaddr == "" {
		if mac := parseHwaddr(ArpFind(ipaddr)); mac != "" {
			hwaddr = mac
			wwlog.Verbose("using %s from arp cache for %s", hwaddr, ipaddr)
		}
	}
	if hwaddr == "" {
		wwlog.Denied("unable to identify node for %s", ipaddr)
		http.Error(w, "unable to identify node", http.StatusUnauthorized)
		return node.Node{}, false
	}

	remoteNode, err := GetNodeOrSetDiscoverable(hwaddr, false)
	if err != nil {
		wwlog.Denied("node not found for hwaddr %s: %s", hwaddr, err)
		http.Error(w, "node not found", http.StatusUnauthorized)
		return node.Node{}, false
	}

	if conf.Warewulf.Secure() && remoteport >= 1024 {
		wwlog.Denied("non-privileged port: %s", req.RemoteAddr)
		http.Error(w, "non-privileged port", http.StatusForbidden)
		return node.Node{}, false
	}

	if remoteNode.AssetKey != "" {
		assetkey := ""
		if len(req.URL.Query()["assetkey"]) > 0 {
			assetkey = req.URL.Query()["assetkey"][0]
		}
		if assetkey == "" {
			wwlog.Denied("missing asset key for node %s", remoteNode.Id())
			http.Error(w, "asset key required", http.StatusUnauthorized)
			return node.Node{}, false
		}
		if assetkey != remoteNode.AssetKey {
			wwlog.Denied("incorrect asset key for node %s", remoteNode.Id())
			http.Error(w, "incorrect asset key", http.StatusForbidden)
			return node.Node{}, false
		}
	}

	return remoteNode, true
}

// HandleFiles serves static files from the configured warewulf files directory.
// Every request must identify a node via ?wwid= or ARP fallback.
// If the node has an asset key, ?assetkey= must match.
// When secure mode is enabled, requests must come from a privileged port.
// If ?render is present, the file is rendered as a Go template for the
// identified node. If the path does not end in .ww but a .ww-suffixed version
// exists, that file is used.
func HandleFiles(w http.ResponseWriter, req *http.Request) {
	conf := warewulfconf.Get()
	filesDir := conf.Paths.WWFilesdir
	wwlog.Debug("Serving file from %s: %s", filesDir, req.URL.Path)

	remoteNode, ok := authenticateNode(w, req)
	if !ok {
		return
	}

	// --- Serve the file ---
	_, render := req.URL.Query()["render"]
	if render {
		if renderName := req.URL.Query().Get("render"); renderName != "" && renderName != remoteNode.Id() {
			http.Error(w, fmt.Sprintf("render node %q does not match identified node %q", renderName, remoteNode.Id()), http.StatusBadRequest)
			return
		}
		filePath := resolveFilesPath(filesDir, req.URL.Path)
		if filePath == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if !strings.HasSuffix(filePath, ".ww") {
			wwFilePath := filePath + ".ww"
			if _, err := os.Stat(wwFilePath); err == nil {
				wwlog.Debug("files: using .ww suffix for %s", filePath)
				filePath = wwFilePath
			} else if _, err := os.Stat(filePath); err == nil {
				http.Error(w, "render requires a .ww template file", http.StatusBadRequest)
				return
			} else {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		} else if _, err := os.Stat(filePath); err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "not found", http.StatusNotFound)
			} else {
				wwlog.Error("files: stat %s: %s", filePath, err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		registry, err := node.New()
		if err != nil {
			wwlog.Error("files: error opening node database: %s", err)
			http.Error(w, fmt.Sprintf("error opening node database: %s", err), http.StatusInternalServerError)
			return
		}

		allNodes, err := registry.FindAllNodes()
		if err != nil {
			wwlog.Error("files: error loading nodes: %s", err)
			http.Error(w, fmt.Sprintf("error loading nodes: %s", err), http.StatusInternalServerError)
			return
		}

		tstruct, err := overlay.InitStruct("", remoteNode, allNodes)
		if err != nil {
			wwlog.Error("files: error initializing template data: %s", err)
			http.Error(w, fmt.Sprintf("error initializing template data: %s", err), http.StatusInternalServerError)
			return
		}
		tstruct.BuildSource = filePath

		buffer, _, _, err := overlay.RenderTemplateFile(filePath, tstruct)
		if err != nil {
			wwlog.Error("files: error rendering template %s: %s", filePath, err)
			http.Error(w, fmt.Sprintf("error rendering template: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
		if _, err := buffer.WriteTo(w); err != nil {
			wwlog.Error("files: error writing response: %s", err)
		}
		wwlog.Info("files: rendered %s for node %s", filePath, remoteNode.Id())
	} else {
		wwlog.Info("files: serving %s for node %s", req.URL.Path, remoteNode.Id())
		fs := noListFileSystem{http.Dir(filesDir)}
		http.StripPrefix("/files/", http.FileServer(fs)).ServeHTTP(w, req)
	}
}
