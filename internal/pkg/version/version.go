package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
)

func GetVersion() string {
	return fmt.Sprintf("%s-%s", buildconfig.VERSION(), buildconfig.RELEASE())
}
