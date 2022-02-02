package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func GetVersion() string {
	return fmt.Sprintf("%s-%s", warewulfconf.Config("VERSION"), warewulfconf.Config("RELEASE"))
}
