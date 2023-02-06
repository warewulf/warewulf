package node

import (
	"fmt"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
)

/*
Checks if for NodeConf all values can be parsed according to their type.
*/
func (nodeConf *NodeConf) Check() (err error) {
	nodeInfoType := reflect.TypeOf(nodeConf)
	nodeInfoVal := reflect.ValueOf(nodeConf)
	// now iterate of every field
	for i := 0; i < nodeInfoVal.Elem().NumField(); i++ {
		//wwlog.Debug("checking field: %s type: %s", nodeInfoType.Elem().Field(i).Name, nodeInfoVal.Elem().Field(i).Type())
		if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.String {
			err = checker(nodeInfoVal.Elem().Field(i).Interface().(string), nodeInfoType.Elem().Field(i).Tag.Get("type"))
			if err != nil {
				return fmt.Errorf("field: %s value:%s err: %s", nodeInfoType.Elem().Field(i).Name, nodeInfoVal.Elem().Field(i).String(), err)
			}
		} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr && !nodeInfoVal.Elem().Field(i).IsNil() {
			nestType := reflect.TypeOf(nodeInfoVal.Elem().Field(i).Interface())
			nestVal := reflect.ValueOf(nodeInfoVal.Elem().Field(i).Interface())
			for j := 0; j < nestType.Elem().NumField(); j++ {
				if nestType.Elem().Field(j).Type.Kind() == reflect.String {
					//wwlog.Debug("checking field: %s type: %s", nestType.Elem().Field(j).Name, nestType.Elem().Field(j).Tag.Get("type"))
					err = checker(nestVal.Elem().Field(j).Interface().(string), nestType.Elem().Field(j).Tag.Get("type"))
					if err != nil {
						return fmt.Errorf("field: %s value:%s err: %s", nestType.Elem().Field(j).Name, nestVal.Elem().Field(j).String(), err)
					}
				}
			}
		} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string]*NetDevs(nil)) {
			netMap := nodeInfoVal.Elem().Field(i).Interface().(map[string]*NetDevs)
			for _, val := range netMap {
				netType := reflect.TypeOf(val)
				netVal := reflect.ValueOf(val)
				for j := 0; j < netType.Elem().NumField(); j++ {
					err = checker(netVal.Elem().Field(j).String(), netType.Elem().Field(j).Tag.Get("type"))
					if err != nil {
						return fmt.Errorf("field: %s value:%s err: %s", netType.Elem().Field(j).Name, netVal.Elem().Field(j).String(), err)
					}
				}
			}
		}
	}
	return nil
}

func checker(value string, valType string) (err error) {
	if valType == "" || value == "" {
		return nil
	}
	//wwlog.Debug("checker: %s is %s", value, valType)
	switch valType {
	case "":
		return nil
	case "bool":
		if strings.ToLower(value) == "yes" {
			return nil
		}
		if strings.ToLower(value) == "no" {
			return nil
		}
		_, err = strconv.ParseBool(value)
		return err
	case "IP":
		_, err = netip.ParseAddr(value)
		return err
	}
	return nil
}
