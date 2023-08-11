package node

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/spf13/cobra"
)

/*
Create cmd line flags from the NodeConf fields. Returns a []func() where every function
must be called, as the commandline parser returns e.g. netip.IP objects which must be parsedf
back to strings.
*/
func (nodeConf *NodeConf) CreateFlags(baseCmd *cobra.Command, excludeList []string) (converters []func() error) {
	/*
		nodeInfoType := reflect.TypeOf(nodeConf)
		nodeInfoVal := reflect.ValueOf(nodeConf)
	*/
	return RecursiveCreateFlags(nodeConf, baseCmd, excludeList)
}

func RecursiveCreateFlags(obj interface{}, baseCmd *cobra.Command, excludeList []string) (converters []func() error) {
	// now iterate of every field
	nodeInfoType := reflect.TypeOf(obj)
	nodeInfoVal := reflect.ValueOf(obj)
	for i := 0; i < nodeInfoVal.Elem().NumField(); i++ {
		if nodeInfoType.Elem().Field(i).Tag.Get("comment") != "" &&
			!util.InSlice(excludeList, nodeInfoType.Elem().Field(i).Tag.Get("lopt")) {
			field := nodeInfoVal.Elem().Field(i)
			converters = append(converters, createFlags(baseCmd, excludeList, nodeInfoType.Elem().Field(i), &field)...)
		} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr {
			newConv := RecursiveCreateFlags(nodeInfoVal.Elem().Field(i).Interface(), baseCmd, excludeList)
			converters = append(converters, newConv...)

		} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Map &&
			nodeInfoType.Elem().Field(i).Type != reflect.TypeOf(map[string]string{}) {
			// add a map with key UNDEF so that it can hold values N.B. UNDEF can never be added through command line
			key := reflect.ValueOf("UNDEF")
			if nodeInfoVal.Elem().Field(i).Len() == 0 {
				if nodeInfoVal.Elem().Field(i).IsNil() {
					nodeInfoVal.Elem().Field(i).Set(reflect.MakeMap(nodeInfoType.Elem().Field(i).Type))
				}
				newPtr := reflect.New(nodeInfoType.Elem().Field(i).Type.Elem().Elem())
				nodeInfoVal.Elem().Field(i).SetMapIndex(key, newPtr)
			} else {
				key = nodeInfoVal.Elem().Field(i).MapKeys()[0]
			}
			newConv := RecursiveCreateFlags(nodeInfoVal.Elem().Field(i).MapIndex(key).Interface(), baseCmd, excludeList)
			converters = append(converters, newConv...)
		}
	}
	return converters
}

/*
Helper function to create the different PersistentFlags() for different types.
*/
func createFlags(baseCmd *cobra.Command, excludeList []string,
	myType reflect.StructField, myVal *reflect.Value) (converters []func() error) {
	if myType.Tag.Get("lopt") != "" {
		if myType.Type.Kind() == reflect.String {
			ptr := myVal.Addr().Interface().(*string)
			switch myType.Tag.Get("type") {
			case "uint":
				converters = append(converters, func() error {
					if !util.InSlice(GetUnsetVerbs(), *ptr) && *ptr != "" {
						_, err := strconv.ParseUint(myType.Tag.Get(*ptr), 10, 32)
						if err != nil {
							return err
						}
					}
					return nil
				})
				if myType.Tag.Get("sopt") != "" {
					baseCmd.PersistentFlags().StringVarP(ptr,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("sopt"),
						myType.Tag.Get("default"),
						myType.Tag.Get("comment"))
				} else {
					baseCmd.PersistentFlags().StringVar(ptr,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("default"),
						myType.Tag.Get("comment"))
				}
			case "bool":
				/*
					Can't use the bool var from pflag as we need the UNSET verbs to be passwd correctly
				*/
				converters = append(converters, func() error {
					if !util.InSlice(GetUnsetVerbs(), *ptr) && *ptr != "" {
						if strings.ToLower(*ptr) == "yes" {
							*ptr = "true"
							return nil
						}
						if strings.ToLower(*ptr) == "no" {
							*ptr = "false"
							return nil
						}
						val, err := strconv.ParseBool(*ptr)
						if err != nil {
							return fmt.Errorf("commandline option %s needs to be bool", myType.Tag.Get("lopt"))
						}
						*ptr = strconv.FormatBool(val)
					}
					return nil
				})
				if myType.Tag.Get("sopt") != "" {
					baseCmd.PersistentFlags().StringVarP(ptr,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("sopt"),
						"",
						myType.Tag.Get("comment"))
				} else {
					baseCmd.PersistentFlags().StringVar(ptr,
						myType.Tag.Get("lopt"),
						"",
						myType.Tag.Get("comment"))
				}
				baseCmd.PersistentFlags().Lookup(myType.Tag.Get("lopt")).NoOptDefVal = "true"
			case "IP":
				converters = append(converters, func() error {
					if !util.InSlice(GetUnsetVerbs(), *ptr) && *ptr != "" {
						ipval := net.ParseIP(*ptr)
						if ipval == nil {
							return fmt.Errorf("commandline option %s needs to be an IP address", myType.Tag.Get("lopt"))
						}
						*ptr = ipval.String()
					}
					return nil
				})
				if myType.Tag.Get("sopt") != "" {
					baseCmd.PersistentFlags().StringVarP(ptr,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("sopt"),
						myType.Tag.Get("default"),
						myType.Tag.Get("comment"))
				} else {
					baseCmd.PersistentFlags().StringVar(ptr,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("default"),
						myType.Tag.Get("comment"))
				}
			case "IPMask":
				defaultConv := net.ParseIP(myType.Tag.Get("default")).DefaultMask()
				var valueRaw net.IPMask
				converters = append(converters, func() error {
					if valueRaw != nil {
						*ptr = valueRaw.String()
						return nil
					} else {
						return fmt.Errorf("could not parse %s to IP", valueRaw.String())
					}
				})
				if myType.Tag.Get("sopt") != "" {
					baseCmd.PersistentFlags().IPMaskVarP(&valueRaw,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("sopt"),
						defaultConv,
						myType.Tag.Get("comment"))
				} else {
					baseCmd.PersistentFlags().IPMaskVar(&valueRaw,
						myType.Tag.Get("lopt"),
						defaultConv,
						myType.Tag.Get("comment"))
				}
			case "MAC":
				converters = append(converters, func() error {
					if !util.InSlice(GetUnsetVerbs(), *ptr) && *ptr != "" {
						myMac, err := net.ParseMAC(*ptr)
						if err != nil {
							return err
						}
						*ptr = myMac.String()
					}
					return nil
				})
				if myType.Tag.Get("sopt") != "" {
					baseCmd.PersistentFlags().StringVarP(ptr,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("sopt"),
						"",
						myType.Tag.Get("comment"))
				} else {
					baseCmd.PersistentFlags().StringVar(ptr,
						myType.Tag.Get("lopt"),
						"",
						myType.Tag.Get("comment"))
				}
			default:
				if myType.Tag.Get("sopt") != "" {
					baseCmd.PersistentFlags().StringVarP(ptr,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("sopt"),
						myType.Tag.Get("default"),
						myType.Tag.Get("comment"))
				} else {
					baseCmd.PersistentFlags().StringVar(ptr,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("default"),
						myType.Tag.Get("comment"))
				}
			}
		} else if myType.Type == reflect.TypeOf([]string{}) {
			ptr := myVal.Addr().Interface().(*[]string)
			if myType.Tag.Get("sopt") != "" {
				baseCmd.PersistentFlags().StringSliceVarP(ptr,
					myType.Tag.Get("lopt"),
					myType.Tag.Get("sopt"),
					[]string{myType.Tag.Get("default")},
					myType.Tag.Get("comment"))
			} else if !util.InSlice(excludeList, myType.Tag.Get("lopt")) {
				baseCmd.PersistentFlags().StringSliceVar(ptr,
					myType.Tag.Get("lopt"),
					[]string{myType.Tag.Get("default")},
					myType.Tag.Get("comment"))

			}
		} else if myType.Type == reflect.TypeOf(map[string]string{}) {
			ptr := myVal.Addr().Interface().(*map[string]string)
			if myType.Tag.Get("sopt") != "" {
				baseCmd.PersistentFlags().StringToStringVarP(ptr,
					myType.Tag.Get("lopt"),
					myType.Tag.Get("sopt"),
					map[string]string{}, // empty default!
					myType.Tag.Get("comment"))
			} else if !util.InSlice(excludeList, myType.Tag.Get("lopt")) {
				baseCmd.PersistentFlags().StringToStringVar(ptr,
					myType.Tag.Get("lopt"),
					map[string]string{}, // empty default!
					myType.Tag.Get("comment"))

			}
		}
	}
	return converters
}
