package warewulfd

import (
	"encoding/json"
	"net/http"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

func sentStatusJSON() ([]byte, error) {
	wwlog.Debug("Request for node sent status data...")

	ret, err := json.MarshalIndent(sentDB, "", "  ")
	if err != nil {
		return ret, errors.Wrap(err, "could not marshal JSON data from status structure")
	}

	return ret, nil

}

func SentStatus(w http.ResponseWriter, req *http.Request) {
	status, err := sentStatusJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(status)
	if err != nil {
		wwlog.Warn("Could not send sent status JSON: %s", err)
	}
}
