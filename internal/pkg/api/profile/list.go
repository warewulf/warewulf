package apiprofile

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
Returns the formatted list of profiles as string
*/
func ProfileList(ShowOpt *wwapiv1.GetProfileList) (profileList wwapiv1.ProfileList, err error) {
	profileList.Output = []string{}
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open node configuration: %s", err)
		return
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Error("Could not find all profiles: %s", err)
		return
	}
	profiles = node.FilterByName(profiles, ShowOpt.Profiles)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Id.Get() < profiles[j].Id.Get()
	})
	if ShowOpt.ShowAll {
		for _, p := range profiles {
			profileList.Output = append(profileList.Output,
				fmt.Sprintf("%s=%s=%s=%s", "PROFILE", "FIELD", "PROFILE", "VALUE"))
			nType := reflect.TypeOf(p)
			nVal := reflect.ValueOf(p)
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
					profileList.Output = append(profileList.Output,
						fmt.Sprintf("%s=%s=%s=%s", p.Id.Print(), fieldName, fieldSource, fieldVal))
				} else if nType.Field(i).Type == reflect.TypeOf(map[string]*node.Entry{}) {
					entrMap := nVal.Field(i).Interface().(map[string]*node.Entry)
					for key, val := range entrMap {
						profileList.Output = append(profileList.Output,
							fmt.Sprintf("%s=%s=%s=%s", p.Id.Print(), key, val.Source(), val.Print()))
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
									profileList.Output = append(profileList.Output,
										fmt.Sprintf("%s=%s=%s=%s", p.Id.Print(), fieldName, fieldSource, fieldVal))
								}
							} else if netInfoType.Field(j).Type == reflect.TypeOf(map[string]*node.Entry{}) {
								for key, val := range netInfoVal.Field(j).Interface().(map[string]*node.Entry) {
									keyfieldName := fieldName + ":" + key
									fieldSource = val.Source()
									fieldVal = val.Print()
									profileList.Output = append(profileList.Output,
										fmt.Sprintf("%s=%s=%s=%s", p.Id.Print(), keyfieldName, fieldSource, fieldVal))
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
							profileList.Output = append(profileList.Output,
								fmt.Sprintf("%s=%s=%s=%s", p.Id.Print(), fieldName, fieldSource, fieldVal))
						} else if nestInfoType.Elem().Field(j).Type == reflect.TypeOf(map[string]*node.Entry{}) {
							for key, val := range nestInfoVal.Elem().Field(j).Interface().(map[string]*node.Entry) {
								fieldName = fieldName + ":" + key
								fieldSource = val.Source()
								fieldVal = val.Print()
								profileList.Output = append(profileList.Output,
									fmt.Sprintf("%s=%s=%s=%s", p.Id.Print(), fieldName, fieldSource, fieldVal))
							}
						}
					}
				}
			}
		}
	} else {
		profileList.Output = append(profileList.Output,
			fmt.Sprintf("%s=%s", "PROFILE NAME", "COMMENT/DESCRIPTION"))

		for _, profile := range profiles {
			profileList.Output = append(profileList.Output,
				fmt.Sprintf("%s=%s", profile.Id.Print(), profile.Comment.Print()))
		}
	}
	return
}
