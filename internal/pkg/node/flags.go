package node

import (
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

/*
Create cmd line flags from the NodeConf fields. Returns a []func() where every function
must be called, as the commandline parser returns e.g. netip.IP objects which must be parsedf
back to strings.
*/
func (nodeConf *NodeConf) CreateFlags(baseCmd *cobra.Command, excludeList []string) (converters []func()) {
	nodeInfoType := reflect.TypeOf(nodeConf)
	nodeInfoVal := reflect.ValueOf(nodeConf)
	// now iterate of every field
	for i := 0; i < nodeInfoVal.Elem().NumField(); i++ {
		if nodeInfoType.Elem().Field(i).Tag.Get("comment") != "" &&
			!util.InSlice(excludeList, nodeInfoType.Elem().Field(i).Tag.Get("lopt")) {
			field := nodeInfoVal.Elem().Field(i)
			converters = append(converters, createFlags(baseCmd, excludeList, nodeInfoType.Elem().Field(i), &field)...)
		} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr {
			nestType := reflect.TypeOf(nodeInfoVal.Elem().Field(i).Interface())
			nestVal := reflect.ValueOf(nodeInfoVal.Elem().Field(i).Interface())
			for j := 0; j < nestType.Elem().NumField(); j++ {
				field := nestVal.Elem().Field(j)
				converters = append(converters, createFlags(baseCmd, excludeList, nestType.Elem().Field(j), &field)...)
			}
		} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string]*NetDevs(nil)) {
			netMap := nodeInfoVal.Elem().Field(i).Interface().(map[string]*NetDevs)
			// add a default network so that it can hold values
			key := "default"
			if len(netMap) == 0 {
				netMap[key] = new(NetDevs)
			} else {
				for keyIt := range netMap {
					key = keyIt
					break
				}
			}
			netType := reflect.TypeOf(netMap[key])
			netVal := reflect.ValueOf(netMap[key])
			for j := 0; j < netType.Elem().NumField(); j++ {
				field := netVal.Elem().Field(j)
				converters = append(converters, createFlags(baseCmd, excludeList, netType.Elem().Field(j), &field)...)
			}
		}
	}
	return converters
}

/*
Helper function to create the different PerisitantFlags() for different types.
*/
func createFlags(baseCmd *cobra.Command, excludeList []string,
	myType reflect.StructField, myVal *reflect.Value) (converters []func()) {
	if myType.Tag.Get("lopt") != "" {
		if myType.Type.Kind() == reflect.String {
			ptr := myVal.Addr().Interface().(*string)
			switch myType.Tag.Get("type") {
			case "uint":
				defaultConv, _ := strconv.ParseUint(myType.Tag.Get("default"), 10, 32)
				var valueRaw uint
				converters = append(converters, func() { *ptr = strconv.FormatUint(uint64(valueRaw), 10) })
				if myType.Tag.Get("sopt") != "" {
					baseCmd.PersistentFlags().UintVarP(&valueRaw,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("sopt"),
						uint(defaultConv),
						myType.Tag.Get("comment"))
				} else {
					baseCmd.PersistentFlags().UintVar(&valueRaw,
						myType.Tag.Get("lopt"),
						uint(defaultConv),
						myType.Tag.Get("comment"))
				}
			case "bool":
				/*
					Can't use the bool var from pflag as we need the UNSET verbs to be passwd correctly
				*/
				converters = append(converters, func() {
					if !util.InSlice(GetUnsetVerbs(), *ptr) && *ptr != "" {
						if strings.ToLower(*ptr) != "yes" {
							*ptr = "true"
							return
						}
						if strings.ToLower(*ptr) != "no" {
							*ptr = "false"
							return
						}
						val, err := strconv.ParseBool(*ptr)
						if err != nil {
							wwlog.Error("commandline option %s needs to be bool", myType.Tag.Get("lopt"))
							os.Exit(1)
						}
						*ptr = strconv.FormatBool(val)
					}
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
				defaultConv := net.ParseIP(myType.Tag.Get("default"))
				var valueRaw net.IP
				converters = append(converters, func() {
					if valueRaw != nil {
						*ptr = valueRaw.String()
					}
				})
				if myType.Tag.Get("sopt") != "" {
					baseCmd.PersistentFlags().IPVarP(&valueRaw,
						myType.Tag.Get("lopt"),
						myType.Tag.Get("sopt"),
						defaultConv,
						myType.Tag.Get("comment"))
				} else {
					baseCmd.PersistentFlags().IPVar(&valueRaw,
						myType.Tag.Get("lopt"),
						defaultConv,
						myType.Tag.Get("comment"))
				}
			case "IPMask":
				defaultConv := net.ParseIP(myType.Tag.Get("default")).DefaultMask()
				var valueRaw net.IPMask
				converters = append(converters, func() { *ptr = valueRaw.String() })
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
