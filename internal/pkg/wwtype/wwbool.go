package wwtype

import (
	"strconv"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/util"
)

/*
Type for holding a simple bool, but which can be set via the UNSET parameter
*/
type WWbool struct {
	bool         // the actual value
	isset   bool // only true if set through yaml or set, so that false can go to disk
	delnext bool // delete after next iteraion
}

// Yaml marshaler, calls this to find out, if going to disk
func (b WWbool) IsZero() bool {
	return !b.isset
}

func (b *WWbool) String() string {
	return strconv.FormatBool(b.bool)
}

func (b *WWbool) Set(str string) error {
	if util.InSlice(GetUnsetVerbs(), str) {
		b.bool = false
		b.isset = true
		b.delnext = true
		return nil
	}

	if strings.ToLower(str) == "yes" || str == "" {
		b.bool = true
		b.isset = true
		return nil
	}
	if strings.ToLower(str) == "no" {
		b.bool = false
		b.isset = true
		return nil
	}
	var err error
	b.bool, err = strconv.ParseBool(str)
	if err == nil {
		b.isset = true
	}
	return err
}

func (b *WWbool) Type() string {
	return "WWbool"
}

func (b WWbool) MarshalBinary() (buf []byte, err error) {
	strconv.AppendBool(buf, b.bool)
	return buf, nil
}

func (b *WWbool) UnmarshalBinary(data []byte) (err error) {
	b.bool, err = strconv.ParseBool(string(data))
	return err
}
func (b WWbool) MarshalText() (buf []byte, err error) {
	if b.bool {
		buf = append(buf, "true"...)
	} else if !b.delnext {
		buf = append(buf, "false"...)
	} else {
		buf = append(buf, "delete"...)
	}
	return buf, nil
}

func (b *WWbool) UnmarshalText(data []byte) (err error) {
	if strings.EqualFold(string(data), "delete") {
		b.isset = false
		b.bool = false
		return nil
	}
	b.isset = true
	return b.Set(string(data))
}
