package version

import (
	"fmt"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

/*
Return the version of wwctl
*/
func Version() string {
	return fmt.Sprintf("%s-%s", warewulfconf.Version, warewulfconf.Release)
}
