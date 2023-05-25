package apinode

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/pkg/hostlist"
)

/*
NodeList lists all to none of the nodes managed by Warewulf. Returns
a formated string slice, with each line as separate string
*/
func NodeList(nodeGet *wwapiv1.GetNodeList) (nodeList wwapiv1.NodeList, err error) {
	// nil is okay for nodeNames
	nodeDB, err := node.New()
	if err != nil {
		return
	}
	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return
	}
	nodeGet.Nodes = hostlist.Expand(nodeGet.Nodes)
	sort.Strings(nodeGet.Nodes)
	if nodeGet.Type == wwapiv1.GetNodeList_Simple {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%s:=:%s:=:%s", "NODE NAME", "PROFILES", "NETWORK"))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			var netNames []string
			for k := range n.NetDevs {
				netNames = append(netNames, k)
			}
			sort.Strings(netNames)
			nodeList.Output = append(nodeList.Output,
				fmt.Sprintf("%s:=:%s:=:%s", n.Id.Print(), n.Profiles.Print(), strings.Join(netNames, ", ")))
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Network {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%s:=:%s:=:%s:=:%s:=:%s:=:%s", "NODE NAME", "NAME", "HWADDR", "IPADDR", "GATEWAY", "DEVICE"))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			if len(n.NetDevs) > 0 {
				for name := range n.NetDevs {
					nodeList.Output = append(nodeList.Output,
						fmt.Sprintf("%s:=:%s:=:%s:=:%s:=:%s:=:%s", n.Id.Print(), name,
							n.NetDevs[name].Hwaddr.Print(),
							n.NetDevs[name].Ipaddr.Print(),
							n.NetDevs[name].Gateway.Print(),
							n.NetDevs[name].Device.Print()))
				}
			} else {
				fmt.Printf("%s:=:%s:=:%s:=:%s:=:%s:=:%s", n.Id.Print(), "--", "--", "--", "--", "--")
			}
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Ipmi {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%s:=:%s:=:%s:=:%s:=:%s", "NODE NAME", "IPMI IPADDR", "IPMI PORT", "IPMI USERNAME", "IPMI INTERFACE"))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			nodeList.Output = append(nodeList.Output,
				fmt.Sprintf("%s:=:%s:=:%s:=:%s:=:%s", n.Id.Print(),
					n.Ipmi.Ipaddr.Print(),
					n.Ipmi.Port.Print(),
					n.Ipmi.UserName.Print(),
					n.Ipmi.Interface.Print()))
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_Long {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%s:=:%s:=:%s:=:%s", "NODE NAME", "KERNEL OVERRIDE", "CONTAINER", "OVERLAYS (S/R)"))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			nodeList.Output = append(nodeList.Output,
				fmt.Sprintf("%s:=:%s:=:%s:=:%s", n.Id.Print(),
					n.Kernel.Override.Print(),
					n.ContainerName.Print(),
					n.SystemOverlay.Print()+"/"+n.RuntimeOverlay.Print()))
		}
	} else if nodeGet.Type == wwapiv1.GetNodeList_All {
		nodeList.Output = append(nodeList.Output,
			fmt.Sprintf("%s:=:%s:=:%s:=:%s", "NODE", "FIELD", "PROFILE", "VALUE"))
		for _, n := range node.FilterByName(nodes, nodeGet.Nodes) {
			nType := reflect.TypeOf(n)
			nVal := reflect.ValueOf(n)
			nConfType := reflect.TypeOf(node.NodeConf{})
			for i := 0; i < nType.NumField(); i++ {
				var fieldName, fieldSource, fieldVal string
				nConfField, ok := nConfType.FieldByName(nType.Field(i).Name)
				if ok {
					fieldName = nConfField.Tag.Get("lopt")
				} else {
					fieldName = nType.Field(i).Name
				}
				if nType.Field(i).Type == reflect.TypeOf(node.Entry{}) {
					entr := nVal.Field(i).Interface().(node.Entry)
					fieldSource = entr.Source()
					fieldVal = entr.Print()
					nodeList.Output = append(nodeList.Output,
						fmt.Sprintf("%s:=:%s:=:%s:=:%s", n.Id.Print(), fieldName, fieldSource, fieldVal))
				} else if nType.Field(i).Type == reflect.TypeOf(map[string]*node.Entry{}) {
					entrMap := nVal.Field(i).Interface().(map[string]*node.Entry)
					for key, val := range entrMap {
						nodeList.Output = append(nodeList.Output,
							fmt.Sprintf("%s:=:%s:=:%s:=:%s", n.Id.Print(), key, val.Source(), val.Print()))
					}
				} else if nType.Field(i).Type == reflect.TypeOf(map[string]*node.NetDevEntry{}) {
					netDevs := nVal.Field(i).Interface().(map[string]*node.NetDevEntry)
					for netName, netWork := range netDevs {
						netInfoType := reflect.TypeOf(*netWork)
						netInfoVal := reflect.ValueOf(*netWork)
						netConfType := reflect.TypeOf(node.NetDevs{})
						for j := 0; j < netInfoType.NumField(); j++ {
							netConfField, ok := netConfType.FieldByName(netInfoType.Field(j).Name)
							if ok {
								if netConfField.Tag.Get("lopt") != "nettagadd" {
									fieldName = netName + ":" + netConfField.Tag.Get("lopt")
								} else {
									fieldName = netName + ":tag"
								}
							} else {
								fieldName = netName + ":" + netInfoType.Field(j).Name
							}
							if netInfoType.Field(j).Type == reflect.TypeOf(node.Entry{}) {
								entr := netInfoVal.Field(j).Interface().(node.Entry)
								fieldSource = entr.Source()
								fieldVal = entr.Print()
								// only print fields with lopt
								if netConfField.Tag.Get("lopt") != "" {
									nodeList.Output = append(nodeList.Output,
										fmt.Sprintf("%s:=:%s:=:%s:=:%s", n.Id.Print(), fieldName, fieldSource, fieldVal))
								}
							} else if netInfoType.Field(j).Type == reflect.TypeOf(map[string]*node.Entry{}) {
								for key, val := range netInfoVal.Field(j).Interface().(map[string]*node.Entry) {
									keyfieldName := fieldName + ":" + key
									fieldSource = val.Source()
									fieldVal = val.Print()
									nodeList.Output = append(nodeList.Output,
										fmt.Sprintf("%s:=:%s:=:%s:=:%s", n.Id.Print(), keyfieldName, fieldSource, fieldVal))
								}
							}

						}
					}
				} else if nType.Field(i).Type.Kind() == reflect.Ptr {
					nestInfoType := reflect.TypeOf(nVal.Field(i).Interface())
					nestInfoVal := reflect.ValueOf(nVal.Field(i).Interface())
					// nestConfType := reflect.TypeOf(nConfField.Type.Elem().FieldByName())
					for j := 0; j < nestInfoType.Elem().NumField(); j++ {
						nestConfField, ok := nConfField.Type.Elem().FieldByName(nestInfoType.Elem().Field(j).Name)
						if ok {
							fieldName = nestConfField.Tag.Get("lopt")
						} else {
							fieldName = nestInfoType.Elem().Field(j).Name
						}
						if nestInfoType.Elem().Field(j).Type == reflect.TypeOf(node.Entry{}) {
							entr := nestInfoVal.Elem().Field(j).Interface().(node.Entry)
							fieldSource = entr.Source()
							fieldVal = entr.Print()
							nodeList.Output = append(nodeList.Output,
								fmt.Sprintf("%s:=:%s:=:%s:=:%s", n.Id.Print(), fieldName, fieldSource, fieldVal))
						} else if nestInfoType.Elem().Field(j).Type == reflect.TypeOf(map[string]*node.Entry{}) {
							for key, val := range nestInfoVal.Elem().Field(j).Interface().(map[string]*node.Entry) {
								fieldName = fieldName + ":" + key
								fieldSource = val.Source()
								fieldVal = val.Print()
								nodeList.Output = append(nodeList.Output,
									fmt.Sprintf("%s:=:%s:=:%s:=:%s", n.Id.Print(), fieldName, fieldSource, fieldVal))
							}
						}
					}
				}

			}
		}
	}
	return
}
