package node

import (
	"reflect"
)

/*
struct to hold the fields of GetFields
*/
type NodeFields struct {
	Field  string
	Source string
	Value  string
}

/*
Get all the info out of NodeInfo. If emptyFields is set true, all fields
are shown not only the ones with effective values
*/
func (node *NodeInfo) GetFields(emptyFields bool) (output []NodeFields) {
	return recursiveFields(node, emptyFields, "")
}

/*
Internal function which travels through all fields of a NodeInfo and for this
reason needs tb called via interface{}
*/
func recursiveFields(obj interface{}, emptyFields bool, prefix string) (output []NodeFields) {
	valObj := reflect.ValueOf(obj)
	typeObj := reflect.TypeOf(obj)
	for i := 0; i < typeObj.Elem().NumField(); i++ {
		if typeObj.Elem().Field(i).Type == reflect.TypeOf(Entry{}) {
			myField := valObj.Elem().Field(i).Interface().(Entry)
			if emptyFields || myField.Get() != "" {
				output = append(output, NodeFields{
					Field:  prefix + typeObj.Elem().Field(i).Name,
					Source: myField.Source(),
					Value:  myField.Print(),
				})
			}
		} else if typeObj.Elem().Field(i).Type == reflect.TypeOf(map[string]*Entry{}) {
			for key, val := range valObj.Elem().Field(i).Interface().(map[string]*Entry) {
				if emptyFields || val.Get() != "" {
					output = append(output, NodeFields{
						Field:  prefix + typeObj.Elem().Field(i).Name + "[" + key + "]",
						Source: val.Source(),
						Value:  val.Print(),
					})
				}
			}
			if valObj.Elem().Field(i).Len() == 0 && emptyFields {
				output = append(output, NodeFields{
					Field: prefix + typeObj.Elem().Field(i).Name + "[]",
				})
			}
		} else if typeObj.Elem().Field(i).Type.Kind() == reflect.Map {
			mapIter := valObj.Elem().Field(i).MapRange()
			for mapIter.Next() {
				nestedOut := recursiveFields(mapIter.Value().Interface(), emptyFields, prefix+typeObj.Elem().Field(i).Name+"["+mapIter.Key().String()+"].")
				if len(nestedOut) == 0 {
					output = append(output, NodeFields{
						Field: prefix + typeObj.Elem().Field(i).Name + "[" + mapIter.Key().String() + "]",
					})
				} else {
					output = append(output, nestedOut...)
				}
			}
			if valObj.Elem().Field(i).Len() == 0 && emptyFields {
				output = append(output, NodeFields{
					Field: prefix + typeObj.Elem().Field(i).Name + "[]",
				})
			}
		} else if typeObj.Elem().Field(i).Type.Kind() == reflect.Ptr {
			nestedOut := recursiveFields(valObj.Elem().Field(i).Interface(), emptyFields, prefix+typeObj.Elem().Field(i).Name)
			output = append(output, nestedOut...)
		}
	}
	return
}

type NodeListEntry interface {
	GetHeader() []string
	GetValue() []string
}

type NodeListResponse struct {
	Nodes map[string][]NodeListEntry `yaml:"Nodes" json:"Nodes"`
}

type NodeListSimpleEntry struct {
	Profile string `yaml:"Profiles" json:"Profiles"`
	Network string `yaml:"Network" json:"Network"`
}

func (n *NodeListSimpleEntry) GetHeader() []string {
	return []string{"NODE NAME", "PROFILES", "NETWORK"}
}

func (n *NodeListSimpleEntry) GetValue() []string {
	return []string{n.Profile, n.Network}
}

type NodeListAllEntry struct {
	Field   string `yaml:"Field" json:"Field"`
	Profile string `yaml:"Profile" json:"Profile"`
	Value   string `yaml:"Value" json:"Value"`
}

func (n *NodeListAllEntry) GetHeader() []string {
	return []string{"NODE", "FIELD", "PROFILE", "VALUE"}
}

func (n *NodeListAllEntry) GetValue() []string {
	return []string{n.Field, n.Profile, n.Value}
}

type NodeListIpmiEntry struct {
	IpmiAddr       string `yaml:"IpmiAddr" json:"IpmiAddr"`
	IpmiPort       string `yaml:"IpmiPort" json:"IpmiPort"`
	IpmiUser       string `yaml:"IpmiUser" json:"IpmiUser"`
	IpmiInterface  string `yaml:"IpmiInterface" json:"IpmiInterface"`
	IpmiEscapeChar string `yaml:"IpmiEscapeChar" json:"IpmiEscapeChar"`
}

func (n *NodeListIpmiEntry) GetHeader() []string {
	return []string{"NODE NAME", "IPMI IPADDR", "IPMI PORT", "IPMI USERNAME", "IPMI INTERFACE", "IPMI ESCAPE CHAR"}
}

func (n *NodeListIpmiEntry) GetValue() []string {
	return []string{n.IpmiAddr, n.IpmiPort, n.IpmiUser, n.IpmiInterface, n.IpmiEscapeChar}
}

type NodeListNetworkEntry struct {
	Name    string `yaml:"Name" json:"Name"`
	HwAddr  string `yaml:"HwAddr" json:"HwAddr"`
	IpAddr  string `yaml:"IpAddr" json:"IpAddr"`
	Gateway string `yaml:"Gateway" json:"Gateway"`
	Device  string `yaml:"Device" json:"Device"`
}

func (n *NodeListNetworkEntry) GetHeader() []string {
	return []string{"NODE NAME", "NAME", "HWADDR", "IPADDR", "GATEWAY", "DEVICE"}
}

func (n *NodeListNetworkEntry) GetValue() []string {
	return []string{n.Name, n.HwAddr, n.IpAddr, n.Gateway, n.Device}
}

type NodeListLongEntry struct {
	KernelOverride string `yaml:"KernelOverride" json:"KernelOverride"`
	Container      string `yaml:"Container" json:"Container"`
	Overlays       string `yaml:"Overlays (S/R)" json:"Overlays (S/R)"`
}

func (n *NodeListLongEntry) GetHeader() []string {
	return []string{"NODE NAME", "KERNEL OVERRIDE", "CONTAINER", "OVERLAYS (S/R)"}
}

func (n *NodeListLongEntry) GetValue() []string {
	return []string{n.KernelOverride, n.Container, n.Overlays}
}
