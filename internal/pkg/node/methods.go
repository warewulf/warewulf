package node

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

/**********
 *
 * Filters
 *
 *********/
/*
Filter a given slice of NodeInfo against a given
regular expression
*/
func FilterByName(set []NodeInfo, searchList []string) []NodeInfo {
	var ret []NodeInfo
	unique := make(map[string]NodeInfo)

	if len(searchList) > 0 {
		for _, search := range searchList {
			for _, entry := range set {
				b, _ := regexp.MatchString("^"+search+"$", entry.Id.Get())
				if b {
					unique[entry.Id.Get()] = entry
				}
			}
		}
		for _, n := range unique {
			ret = append(ret, n)
		}
	} else {
		ret = set
	}

	return ret
}

/**********
 *
 * Sets
 *
 *********/

/*
 Set value. If argument is 'UNDEF', 'DELETE',
 'UNSET" or '--' the value is removed.
 N.B. the '--' might never ever happen as '--'
 is parsed out by cobra
*/
func (ent *Entry) Set(val string) {
	if val == "" {
		return
	}

	if val == "UNDEF" || val == "DELETE" || val == "UNSET" || val == "--" || val == "nil" {
		wwlog.Debug("Removing value for %v\n", *ent)
		ent.value = []string{""}
	} else {
		ent.value = []string{val}
	}
}

/*
Set bool
*/
func (ent *Entry) SetB(val bool) {
	if val {
		ent.value = []string{"true"}
	} else {
		ent.value = []string{"false"}
	}
}

func (ent *Entry) SetSlice(val []string) {
	if len(val) == 0 {
		return
	}
	if val[0] == "UNDEF" || val[0] == "DELETE" || val[0] == "UNSET" || val[0] == "--" {
		ent.value = []string{}
	} else {
		ent.value = val
	}
}

/*
Set alternative value
*/
func (ent *Entry) SetAlt(val string, from string) {
	if val == "" {
		return
	}
	ent.altvalue = []string{val}
	ent.from = from
}

/*
Sets alternative bool
*/
func (ent *Entry) SetAltB(val bool, from string) {
	if val {
		ent.altvalue = []string{"true"}
		ent.from = from
	} else {
		ent.altvalue = []string{"false"}
		ent.from = from
	}
}

/*
Sets alternative slice
*/
func (ent *Entry) SetAltSlice(val []string, from string) {
	if len(val) == 0 {
		return
	}
	ent.altvalue = val
	ent.from = from
}

/*
Sets the default value of an entry.
*/
func (ent *Entry) SetDefault(val string) {
	if val == "" {
		return
	}
	ent.def = []string{val}

}

/*
Set the default entry as slice
*/
func (ent *Entry) SetDefaultSlice(val []string) {
	if len(val) == 0 {
		return
	}
	ent.def = val

}

/**********
*
* Gets
*
*********/
/*
Gets the the entry of the value in folowing order
* node value if set
* profile value if set
* default value if set
*/
func (ent *Entry) Get() string {
	if len(ent.value) != 0 {
		return ent.value[0]
	}
	if len(ent.altvalue) != 0 {
		return ent.altvalue[0]
	}
	if len(ent.def) != 0 {
		return ent.def[0]
	}
	return ""
}

/*
Get the bool value of an entry.
*/
func (ent *Entry) GetB() bool {
	if len(ent.value) == 0 || ent.value[0] == "false" || ent.value[0] == "no" {
		if len(ent.altvalue) == 0 || ent.altvalue[0] == "false" || ent.altvalue[0] == "no" {
			return false
		}
		return false
	}
	return true
}

/*
Returns a string slice created from a comma seperated list of the value.
*/
func (ent *Entry) GetSlice() []string {
	var retval []string
	if len(ent.value) != 0 {
		return ent.value
	}
	if len(ent.altvalue) != 0 {
		return ent.altvalue
	}
	if len(ent.def) != 0 {
		return ent.def
	}
	return retval
}

/*
Get the real value, not the alternative of default one.
*/
func (ent *Entry) GetReal() string {
	if len(ent.value) == 0 {
		return ""
	}
	return ent.value[0]
}

/*
Get the real value, not the alternative of default one.
*/
func (ent *Entry) GetRealSlice() []string {
	if len(ent.value) == 0 {
		return []string{}
	}
	return ent.value
}

/*
true if the entry has set a real value, else false.
*/
func (ent *Entry) GotReal() bool {
	return len(ent.value) != 0
}

/**********
 *
 * Misc
 *
 *********/

/*
Returns the value of Entry if it was defined set or
alternative is presend. Default value is in '()'. If
nothing is defined '--' is returned.
*/
func (ent *Entry) Print() string {
	if len(ent.value) != 0 {
		return strings.Join(ent.value, ",")
	}
	if len(ent.altvalue) != 0 {
		return strings.Join(ent.altvalue, ",")
	}
	if len(ent.def) != 0 {
		return "(" + strings.Join(ent.def, ",") + ")"
	}
	return "--"
}

/*
Was used for combined stringSlice

func (ent *Entry) PrintComb() string {
	if ent.value != "" && ent.altvalue != "" {
		return "[" + ent.value + "," + ent.altvalue + "]"
	}
	return ent.Print()
}
*/

/*
same as GetB()
*/
func (ent *Entry) PrintB() string {
	if len(ent.value) != 0 || len(ent.altvalue) != 0 {
		return fmt.Sprintf("%t", ent.GetB())
	}
	return fmt.Sprintf("(%t)", ent.GetB())
}

/*
Returns SUPERSEDED if value was set per node or
per profile. Else -- is returned.
*/
func (ent *Entry) Source() string {
	if len(ent.value) != 0 && len(ent.altvalue) != 0 {
		return "SUPERSEDED"
		//return fmt.Sprintf("[%s]", ent.from)
	} else if ent.from == "" {
		return "--"
	}
	return ent.from
}

/*
Check if value was defined.
*/
func (ent *Entry) Defined() bool {
	if len(ent.value) != 0 {
		return true
	}
	if len(ent.altvalue) != 0 {
		return true
	}
	if len(ent.def) != 0 {
		return true
	}
	return false
}

/*
Set the Entry trough an interface by trying to cast the interface
*/
func SetEntry(entryPtr interface{}, val interface{}) {
	if reflect.TypeOf(entryPtr) == reflect.TypeOf((*Entry)(nil)) {
		entry := entryPtr.(*Entry)
		valKind := reflect.TypeOf(val)
		if valKind.Kind() == reflect.String {
			entry.Set(val.(string))
		} else if valKind.Kind() == reflect.Slice {
			if valKind.Elem().Kind() == reflect.String {
				entry.SetSlice(val.([]string))
			} else {
				panic("Got unknown slice type")
			}
		}
	} else {
		panic(fmt.Sprintf("Can't convert %s to *node.Entry\n", reflect.TypeOf(entryPtr)))
	}

}

/*
Call SetEntry for given field (NodeInfo)
*/
func (node *NodeInfo) SetField(fieldName string, val interface{}) {
	field := reflect.ValueOf(node).Elem().FieldByName(fieldName)
	if field.IsValid() {
		//fmt.Println(reflect.TypeOf(field.Addr().Interface()))
		SetEntry(field.Addr().Interface(), val)
	} else {
		fieldNames := strings.Split(fieldName, ".")
		if len(fieldNames) == 2 {
			nestedField := reflect.ValueOf(node).Elem().FieldByName(fieldNames[0])
			if nestedField.IsValid() {
				switch nestedField.Addr().Type() {
				case reflect.TypeOf((**KernelEntry)(nil)):
					entry := nestedField.Addr().Interface().(**KernelEntry)
					(*entry).SetField(fieldNames[1], val)
				case reflect.TypeOf((**IpmiEntry)(nil)):
					entry := nestedField.Addr().Interface().(**IpmiEntry)
					(*entry).SetField(fieldNames[1], val)
				default:
					panic(fmt.Sprintf("not implemented type %v\n", nestedField.Addr().Type()))
				}
			} else {
				panic(fmt.Sprintf("field %s is not a nested type of %s", fieldNames[0], fieldName))
			}
		} else {
			panic(fmt.Sprintf("field %s does not exists in node.NodeInfo\n", fieldName))
		}
	}

}

/*
Call SetEntry for given field (KernelEntry)
*/
func (node *KernelEntry) SetField(fieldName string, val interface{}) {
	field := reflect.ValueOf(node).Elem().FieldByName(fieldName)
	if field.IsValid() {
		SetEntry(field.Addr().Interface(), val)
	} else {
		panic(fmt.Sprintf("field %s does not exists in node.KernEntry\n", fieldName))
	}
}

/*
Call SetEntry for given field (ImpiEntry)
*/
func (node *IpmiEntry) SetField(fieldName string, val interface{}) {
	field := reflect.ValueOf(node).Elem().FieldByName(fieldName)
	if field.IsValid() {
		SetEntry(field.Addr().Interface(), val)
	} else {
		panic(fmt.Sprintf("field %s does not exists in node.KernEntry\n", fieldName))
	}
}

/*
Get all names of the fields in the given struct (recursive)
and create a map[name of struct field]*string if the the field
of the struct bears the comment tag.
*/
func GetOptionsMap(theStruct interface{}) map[string]*string {
	optionsMap := make(map[string]*string)
	structVal := reflect.ValueOf(theStruct)
	structTyp := structVal.Type()
	for i := 0; i < structVal.NumField(); i++ {
		field := structTyp.Field(i)
		if field.Type.Kind() == reflect.Struct {
			subStruct := GetOptionsMap(field)
			for key, val := range subStruct {
				optionsMap[key] = val
			}
		} else if field.Tag.Get("comment") != "" {
			var newStr string
			optionsMap[field.Name] = &newStr
		}

	}
	return optionsMap
}

type CobraCommand struct {
	*cobra.Command
}

/*
Get all names of the fields in the given struct (recursive)
and create a map[name of struct field]*string if the the field
of the struct bears the comment tag.
*/
func (baseCmd *CobraCommand) CreateFlags(theStruct interface{}) map[string]*string {
	optionsMap := make(map[string]*string)
	structVal := reflect.ValueOf(theStruct)
	structTyp := structVal.Type()
	for i := 0; i < structVal.NumField(); i++ {
		field := structTyp.Field(i)
		fmt.Printf("%s: field.Kind() == %s\n", field.Name, field.Type.Kind())
		if field.Type.Kind() == reflect.Ptr {
			a := structVal.Field(i).Elem().Interface()
			subStruct := baseCmd.CreateFlags(a)
			for key, val := range subStruct {
				optionsMap[field.Name+"."+key] = val
			}

		} else if field.Type.Kind() == reflect.Map {
			// Just check for network map
			fmt.Println(reflect.TypeOf(structVal.Field(i).Elem()))

		} else if field.Tag.Get("comment") != "" {
			var newStr string
			optionsMap[field.Name] = &newStr
			if field.Tag.Get("sopt") != "" {
				baseCmd.PersistentFlags().StringVarP(&newStr,
					field.Tag.Get("lopt"),
					field.Tag.Get("sopt"),
					field.Tag.Get("default"),
					field.Tag.Get("comment"))
			} else {
				baseCmd.PersistentFlags().StringVar(&newStr,
					field.Tag.Get("lopt"),
					field.Tag.Get("default"),
					field.Tag.Get("comment"))

			}
		}

	}
	return optionsMap
}
