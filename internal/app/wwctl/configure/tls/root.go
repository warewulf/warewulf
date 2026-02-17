package tls

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "tls [OPTIONS]",
		Aliases:               []string{"keys", "key", "cert", "crt"},
		Short:                 "Manage and initialize x509 keys",
		Long:                  `This application allows you to manage the x509 keys and certificates for Warewulf.`,
		RunE:                  CobraRunE,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     completions.None,
	}
	importPath string
	exportPath string
	create     bool
	force      bool
)

func init() {
	baseCmd.PersistentFlags().StringVar(&importPath, "import", "", "Import keys from directory")
	baseCmd.PersistentFlags().StringVar(&exportPath, "export", "", "Export keys to directory")
	baseCmd.PersistentFlags().BoolVar(&create, "create", false, "Create keys if they do not exist")
	baseCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Enforce creation of keys even if they exist")
}

// GetCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
