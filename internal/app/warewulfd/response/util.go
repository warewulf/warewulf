package response

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func getSanity(req *http.Request) (node.NodeInfo, error) {
	url := strings.Split(req.URL.Path, "/")
	var ret node.NodeInfo

	nodes, err := node.New()
	if err != nil {
		return ret, errors.New(fmt.Sprintf("%s", err))
	}

	hwaddr := strings.ReplaceAll(url[2], "-", ":")
	ret, err = nodes.FindByHwaddr(hwaddr)
	if err != nil {
		return ret, errors.New("Could not find node by HW address")
	}

	if ret.Fqdn == "" {
		log.Printf("UNKNOWN: %15s: %s\n", hwaddr, req.URL.Path)
		return ret, errors.New("Unknown node HW address: " + hwaddr)
	} else {
		log.Printf("REQ:   %15s: %s\n", ret.Fqdn, req.URL.Path)
	}

	return ret, nil
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
