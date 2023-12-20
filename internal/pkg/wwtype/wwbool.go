package wwtype

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/util"
)

// Simple string which can be converted to bool. Backend storage
// is string for better merging
type WWbool string

/*
Transform the underlying string value to bool
*/
func (val WWbool) Bool() bool {
	str := string(val)
	if util.InSlice(GetUnsetVerbs(), str) {
		return false
	}
	if strings.ToLower(str) == "yes" {
		return true
	}
	if strings.ToLower(str) == "no" {
		return false
	}
	bval, _ := strconv.ParseBool(str)
	return bval
}

/*
Set the string, only accept bool values like true, false, but also UNDEF
*/
func (val *WWbool) Set(str string) error {
	if util.InSlice(GetUnsetVerbs(), str) {
		// run the unset verb trough, will be filtered out later
		*val = WWbool(str)
		return nil
	}
	if strings.ToLower(str) == "yes" {
		*val = WWbool("true")
		return nil
	}
	if strings.ToLower(str) == "no" {
		*val = WWbool("false")
		return nil
	}
	bval, err := strconv.ParseBool(str)
	if err == nil {
		*val = WWbool(strconv.FormatBool(bval))
		return nil
	}
	return fmt.Errorf("value for WWbool can't be set from %s", str)
}

func (val WWbool) String() string {
	return string(val)
}

func (b WWbool) Type() string {
	return "WWbool"
}
