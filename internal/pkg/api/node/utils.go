package apinode

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

// nodeDbSave persists the nodeDB to disk and restarts warewulfd.
// TODO: We will likely need locking around anything changing nodeDB
// or restarting warewulfd. Determine if the reason for restart is
// just to reinitialize warewulfd with the new nodeDB or if there is
// something more to it.
func DbSave(nodeDB *node.NodeYaml) (err error) {
	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist nodedb")
	}

	err = warewulfd.DaemonReload()
	if err != nil {
		return errors.Wrap(err, "failed to reload warewulf daemon")
	}
	return
}

/*
Add the netname to the options map, as its only known after the
command line options have been read out, but its needing for setting
the values. Returns the manipulated map and a bool which is true if
networks specific values were set, but no netname was given.
*/
func AddNetname(theMap map[string]*string) (map[string]*string, bool) {
	netname := ""
	netvalues := false
	retMap := make(map[string]*string)
	for key, val := range theMap {
		if key == "NetDevs" && *val != "" {
			netname = *val
		}
	}
	for key, val := range theMap {
		keys := strings.Split(key, ".")
		myVal := *val
		if len(keys) >= 2 && keys[0] == "NetDevs" {
			if *val != "" {
				netvalues = true
			}
			if netname != "" {
				retMap[keys[0]+"."+netname+"."+strings.Join(keys[1:], ".")] = &myVal
			}
		} else {
			retMap[key] = &myVal
		}

	}

	return retMap, netvalues && netname == ""
}
