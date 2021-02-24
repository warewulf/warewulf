package warewulfd

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/node"
)

func getSanity(req *http.Request) (node.NodeInfo, error) {
	url := strings.Split(req.URL.Path, "/")

	hwaddr := strings.ReplaceAll(url[2], "-", ":")

	nodeobj, err := GetNode(hwaddr)
	if err != nil {
		var ret node.NodeInfo
		return ret, errors.New("Could not find node by HW address: " + req.URL.Path)
	}

	log.Printf("REQ:   %15s: %s\n", nodeobj.Id.Get(), req.URL.Path)

	return nodeobj, nil
}

func sendFile(w http.ResponseWriter, filename string, sendto string) error {

	fd, err := os.Open(filename)
	if err != nil {
		return err
	}

	FileHeader := make([]byte, 512)
	fd.Read(FileHeader)
	FileContentType := http.DetectContentType(FileHeader)
	FileStat, _ := fd.Stat()
	FileSize := strconv.FormatInt(FileStat.Size(), 10)

	w.Header().Set("Content-Disposition", "attachment; filename=kernel")
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	fd.Seek(0, 0)
	io.Copy(w, fd)

	fd.Close()
	return nil
}
