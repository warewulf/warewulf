package wwtype

import (
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/util"
)

func GetUnsetVerbs() []string {
	return []string{"unset", "delete", "undef", "--", "nil", "0.0.0.0"}
}

func IsUnsetVerb(value string) bool {
	return util.InSlice(GetUnsetVerbs(), strings.ToLower(value))
}
