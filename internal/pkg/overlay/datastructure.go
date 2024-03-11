package overlay

import (
	"bytes"
	"encoding/gob"
	"os"
	"strconv"
	"time"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

/*
struct which contains the variables to which are available in
the templates.
*/
type TemplateStruct struct {
	Id            string
	Hostname      string
	BuildHost     string
	BuildTime     string
	BuildTimeUnix string
	BuildSource   string
	Ipaddr        string
	Ipaddr6       string
	Netmask       string
	Network       string
	NetworkCIDR   string
	Ipv6          bool
	Dhcp          warewulfconf.DHCPConf
	Nfs           warewulfconf.NFSConf
	Warewulf      warewulfconf.WarewulfConf
	Tftp          warewulfconf.TFTPConf
	Paths         warewulfconf.BuildConfig
	AllNodes      []node.NodeConf
	node.NodeConf
	// backward compatiblity
	Container string
	ThisNode  *node.NodeConf
}

/*
Initialize an TemplateStruct with the given node.NodeInfo
*/
func InitStruct(nodeData node.NodeConf) (TemplateStruct, error) {
	var tstruct TemplateStruct
	hostname, _ := os.Hostname()
	tstruct.BuildHost = hostname
	// controller := warewulfconf
	nodeDB, err := node.New()
	if err != nil {
		return tstruct, err
	}
	tstruct.ThisNode = &nodeData
	allNodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return tstruct, err
	}
	// init some convenience vars
	tstruct.Id = nodeData.Id()
	tstruct.Hostname = nodeData.Id()
	// Backwards compatibility for templates using "Keys"
	tstruct.AllNodes = allNodes
	dt := time.Now()
	tstruct.BuildTime = dt.Format("01-02-2006 15:04:05 MST")
	tstruct.BuildTimeUnix = strconv.FormatInt(dt.Unix(), 10)
	tstruct.NodeConf.Tags = map[string]string{}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err = enc.Encode(nodeData)
	if err != nil {
		return tstruct, err
	}
	err = dec.Decode(&tstruct)
	if err != nil {
		return tstruct, err
	}
	return tstruct, nil

}
