package version

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Makefile modifies version and release during build using VERSION and RELEASE.
// @see https://polyverse.com/blog/how-to-embed-versioning-information-in-go-applications-f76e2579b572/
var (
	version = "tbs"
	release = "tbs"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	fmt.Printf("v%s.%s\n", version, release)
	return nil
}
