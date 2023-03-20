package node

import (
	"reflect"

	"github.com/spf13/cobra"

	"github.com/hpcng/warewulf/internal/pkg/util"
)

/*
Create cmd line flags from the NodeConf fields
*/
func (nodeConf *NodeConf) CreateFlags(baseCmd *cobra.Command, excludeList []string) {
	nodeInfoType := reflect.TypeOf(nodeConf)
	nodeInfoVal := reflect.ValueOf(nodeConf)
	// now iterate of every field
	for i := 0; i < nodeInfoVal.Elem().NumField(); i++ {
		if nodeInfoType.Elem().Field(i).Tag.Get("comment") != "" &&
			!util.InSlice(excludeList, nodeInfoType.Elem().Field(i).Tag.Get("lopt")) {
			field := nodeInfoVal.Elem().Field(i)
			createFlags(baseCmd, excludeList, nodeInfoType.Elem().Field(i), &field)
		} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr {
			nestType := reflect.TypeOf(nodeInfoVal.Elem().Field(i).Interface())
			nestVal := reflect.ValueOf(nodeInfoVal.Elem().Field(i).Interface())
			for j := 0; j < nestType.Elem().NumField(); j++ {
				field := nestVal.Elem().Field(j)
				createFlags(baseCmd, excludeList, nestType.Elem().Field(j), &field)
			}
		} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string]*NetDevConf(nil)) {
			netMap := nodeInfoVal.Elem().Field(i).Interface().(map[string]*NetDevConf)
			// add a default network so that it can hold values
			key := "default"
			if len(netMap) == 0 {
				netMap[key] = new(NetDevConf)
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
				createFlags(baseCmd, excludeList, netType.Elem().Field(j), &field)
			}
		}
	}
}

/*
Helper function to create the different PerisitantFlags() for different types.
*/
func createFlags(baseCmd *cobra.Command, excludeList []string,
	myType reflect.StructField, myVal *reflect.Value) {
	if myType.Tag.Get("lopt") != "" {
		if myType.Type.Kind() == reflect.String {
			ptr := myVal.Addr().Interface().(*string)
			if myType.Tag.Get("sopt") != "" {
				baseCmd.PersistentFlags().StringVarP(ptr,
					myType.Tag.Get("lopt"),
					myType.Tag.Get("sopt"),
					myType.Tag.Get("default"),
					myType.Tag.Get("comment"))
			} else if !util.InSlice(excludeList, myType.Tag.Get("lopt")) {
				baseCmd.PersistentFlags().StringVar(ptr,
					myType.Tag.Get("lopt"),
					myType.Tag.Get("default"),
					myType.Tag.Get("comment"))

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
}
