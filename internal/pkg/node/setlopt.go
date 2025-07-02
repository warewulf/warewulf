package node

import (
	"net"
	"reflect"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"github.com/warewulf/warewulf/internal/pkg/wwtype"
)

/*
Set the field of the NodeConf with the given lopt name, returns true if the
field was found. String slices must be comma separated. Network must have the form
net.$NETNAME.lopt or netname.$NETNAME.lopt
*/
func (node *Node) SetLopt(lopt string, value string) (found bool) {
	found = false
	nodeType := reflect.TypeOf(node)
	nodeVal := reflect.ValueOf(node)

	// node fields
	for i := range nodeVal.Elem().NumField() {
		fieldLopt := nodeType.Elem().Field(i).Tag.Get("lopt")
		if fieldLopt == lopt {
			fieldType := nodeType.Elem().Field(i).Type
			if fieldType == reflect.TypeOf(wwtype.WWbool("")) {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s",
					lopt, nodeType.Elem().Field(i).Name, value)
				confVal := nodeVal.Elem().Field(i).Addr().Interface().(*wwtype.WWbool)
				*confVal = wwtype.WWbool(value)
			} else if fieldType.Kind() == reflect.String {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s",
					lopt, nodeType.Elem().Field(i).Name, value)
				confVal := nodeVal.Elem().Field(i).Addr().Interface().(*string)
				*confVal = value
				found = true
			} else if fieldType == reflect.TypeOf([]string{}) {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s",
					lopt, nodeType.Elem().Field(i).Name, value)
				confVal := nodeVal.Elem().Field(i).Addr().Interface().(*[]string)
				*confVal = strings.Split(value, ",")
				found = true
			}
		}
	}

	// profile fields
	profileType := reflect.TypeOf(&node.Profile)
	profileVal := reflect.ValueOf(&node.Profile)
	for i := range profileVal.Elem().NumField() {
		fieldLopt := profileType.Elem().Field(i).Tag.Get("lopt")
		if fieldLopt == lopt {
			fieldType := profileType.Elem().Field(i).Type
			if fieldType == reflect.TypeOf(wwtype.WWbool("")) {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s",
					lopt, profileType.Elem().Field(i).Name, value)
				confVal := profileVal.Elem().Field(i).Addr().Interface().(*wwtype.WWbool)
				*confVal = wwtype.WWbool(value)
			} else if fieldType.Kind() == reflect.String {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s",
					lopt, profileType.Elem().Field(i).Name, value)
				confVal := profileVal.Elem().Field(i).Addr().Interface().(*string)
				*confVal = value
				found = true
			} else if fieldType == reflect.TypeOf([]string{}) {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s",
					lopt, profileType.Elem().Field(i).Name, value)
				confVal := profileVal.Elem().Field(i).Addr().Interface().(*[]string)
				*confVal = strings.Split(value, ",")
				found = true
			}
		}
	}

	// netdev fields
	loptSlice := strings.Split(lopt, ".")
	wwlog.Debug("Trying to get network out of %s", loptSlice)
	if !found && len(loptSlice) == 3 && (loptSlice[0] == "net" || loptSlice[0] == "network" || loptSlice[0] == "netname") {
		if node.NetDevs == nil {
			node.NetDevs = make(map[string]*NetDev)
		}
		if node.NetDevs[loptSlice[1]] == nil {
			node.NetDevs[loptSlice[1]] = new(NetDev)
		}
		netDevType := reflect.TypeOf(node.NetDevs[loptSlice[1]])
		netDevVal := reflect.ValueOf(node.NetDevs[loptSlice[1]])
		for i := range netDevVal.Elem().NumField() {
			if netDevType.Elem().Field(i).Tag.Get("lopt") == loptSlice[2] {
				fieldType := netDevType.Elem().Field(i).Type
				if fieldType == reflect.TypeOf(wwtype.WWbool("")) {
					wwlog.Verbose("Found lopt %s for network %s mapping to %s, setting to %s",
						lopt, loptSlice[1], netDevType.Elem().Field(i).Name, value)
					confVal := netDevVal.Elem().Field(i).Addr().Interface().(*wwtype.WWbool)
					*confVal = wwtype.WWbool(value)
					found = true
				} else if fieldType == reflect.TypeOf(net.IP{}) {
					wwlog.Verbose("Found lopt %s for network %s mapping to %s, setting to %s",
						lopt, loptSlice[1], netDevType.Elem().Field(i).Name, value)
					confVal := netDevVal.Elem().Field(i).Addr().Interface().(*net.IP)
					*confVal = net.ParseIP(value)
				} else if fieldType.Kind() == reflect.String {
					wwlog.Verbose("Found lopt %s for network %s mapping to %s, setting to %s",
						lopt, loptSlice[1], netDevType.Elem().Field(i).Name, value)
					confVal := netDevVal.Elem().Field(i).Addr().Interface().(*string)
					*confVal = value
					found = true
				} else if fieldType == reflect.TypeOf([]string{}) {
					wwlog.Verbose("Found lopt %s for network %s mapping to %s, setting to %s",
						lopt, loptSlice[1], netDevType.Elem().Field(i).Name, value)
					confVal := netDevVal.Elem().Field(i).Addr().Interface().(*[]string)
					*confVal = strings.Split(value, ",")
					found = true
				}
			}
		}
	}
	return found
}
