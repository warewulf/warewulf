package wwtype

import (
	"fmt"
	"strconv"
	"strings"
)

// Simple string which can be converted to bool. Backend storage
// is string for better merging
type WWbool string

/*
Transform the underlying string value to bool
*/
func (val WWbool) Bool() bool {
	str := strings.ToLower(string(val))
	if IsUnsetVerb(str) {
		return false
	}
	switch str {
	case "yes":
		return true
	case "no", "":
		return false
	}
	bval, err := strconv.ParseBool(str)
	if err != nil {
		return false
	}
	return bval
}

func (val WWbool) BoolDefaultTrue() bool {
	str := strings.ToLower(string(val))
	if IsUnsetVerb(str) {
		return false
	}
	switch str {
	case "yes", "":
		return true
	case "no":
		return false
	}
	bval, err := strconv.ParseBool(str)
	if err != nil {
		return true
	}
	return bval
}

/*
Set the string, only accept bool values like true, false, but also UNDEF
*/
func (val *WWbool) Set(str string) error {
	if IsUnsetVerb(str) {
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
