package list

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

type variables struct {
	full       bool
	size       bool
	kernel     bool
	chroot     bool
	compressed bool
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS]",
		Short:                 "List imported Warewulf images",
		Long:                  "This command will show you the images that are imported into Warewulf.",
		RunE:                  CobraRunE(&vars),
		Aliases:               []string{"ls"},
		ValidArgsFunction:     completions.Images,
	}
	baseCmd.PersistentFlags().BoolVarP(&vars.full, "long", "l", false, "show all")
	baseCmd.PersistentFlags().BoolVarP(&vars.kernel, "kernel", "k", false, "show kernel version")
	baseCmd.PersistentFlags().BoolVarP(&vars.size, "size", "s", false, "show size information")
	baseCmd.PersistentFlags().BoolVarP(&vars.chroot, "chroot", "c", false, "show size of chroot")
	baseCmd.PersistentFlags().BoolVar(&vars.compressed, "compressed", false, "show size of the compressed image")
	return baseCmd
}
