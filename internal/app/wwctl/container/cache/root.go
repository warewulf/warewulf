package ocicache

import (
	"github.com/spf13/cobra"
	cachelist "github.com/warewulf/warewulf/internal/app/wwctl/container/cache/clean"
	cacheclean "github.com/warewulf/warewulf/internal/app/wwctl/container/cache/list"
)

var baseCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   "cache COMMAND [OPTIONS]",
	Short:                 "Manage the cached blobs",
	Long: `When importing a container from a registry the so called blobs
re written to cache on disk. This command allows to show and
mange these blobs.`,
}

func init() {
	baseCmd.AddCommand(cachelist.GetCommand())
	baseCmd.AddCommand(cacheclean.GetCommand())
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
