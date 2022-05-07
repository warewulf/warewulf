package warewulfd

import (
	"io"
	"bytes"
	"net/http"
	"os"
	"strconv"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

func sendFile(w http.ResponseWriter, filename string, sendto string) error {
	fd, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer fd.Close()

	FileHeader := make([]byte, 512)
	_, err = fd.Read(FileHeader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.Wrap(err, "failed to read header")
	}

	FileContentType := http.DetectContentType(FileHeader)
	FileStat, _ := fd.Stat()
	FileSize := strconv.FormatInt(FileStat.Size(), 10)

	_, err = fd.Seek(0, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.Wrap(err, "failed to seek")
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, fd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.Wrap(err, "failed to copy")
	}

	w.Header().Set("Content-Disposition", "attachment; filename=kernel")
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)
	buf.WriteTo(w)

	wwlog.Info("SEND:  %15s: %s", sendto, filename)

	return nil
}
