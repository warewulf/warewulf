package warewulfd

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/pkg/errors"
)

var logwriter *syslog.Writer
var loginit bool

func daemonLogf(message string, a ...interface{}) {
	conf, err := warewulfconf.New()
	if err != nil {
		fmt.Printf("ERROR: Could not read Warewulf configuration file: %s\n", err)
		return
	}

	if conf.Warewulf.Syslog {
		if !loginit {
			var err error

			logwriter, err = syslog.New(syslog.LOG_NOTICE, "warewulfd")
			if err != nil {
				return
			}
			log.SetOutput(logwriter)
			loginit = true
		}

		log.SetFlags(0)
		log.SetPrefix("")
		log.Printf(message, a...)

	} else {
		prefix := fmt.Sprintf("[%s] ", time.Now().Format(time.UnixDate))
		fmt.Printf(prefix+message, a...)
	}
}

func getSanity(req *http.Request) (node.NodeInfo, error) {
	url := strings.Split(req.URL.Path, "/")

	hwaddr := strings.ReplaceAll(url[2], "-", ":")

	nodeobj, err := GetNode(hwaddr)
	if err != nil {
		var ret node.NodeInfo
		return ret, errors.New("Could not find node by HW address: " + req.URL.Path)
	}

	daemonLogf("REQ:   %15s: %s\n", nodeobj.Id.Get(), req.URL.Path)

	return nodeobj, nil
}

func sendFile(w http.ResponseWriter, filename string, sendto string) error {
	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	FileHeader := make([]byte, 512)
	_, err = fd.Read(FileHeader)
	if err != nil {
		return errors.Wrap(err, "failed to read header")
	}
	FileContentType := http.DetectContentType(FileHeader)
	FileStat, _ := fd.Stat()
	FileSize := strconv.FormatInt(FileStat.Size(), 10)

	w.Header().Set("Content-Disposition", "attachment; filename=kernel")
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	_, err = fd.Seek(0, 0)
	if err != nil {
		return errors.Wrap(err, "failed to seek")
	}

	_, err = io.Copy(w, fd)
	if err != nil {
		return errors.Wrap(err, "failed to copy")
	}

	daemonLogf("SEND:  %15s: %s\n", sendto, filename)

	return nil
}
