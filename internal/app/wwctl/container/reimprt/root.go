package reimprt

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
)

type variables struct {
	fromCache   bool
	syncUser    bool
	setBuild    bool
	ociNoHttps  bool
	ociUsername string
	ociPassword string
}

func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "reimport CONTAINER NEWCONTAINER",
		Short:                 "Reimport a container",
		Long:                  "Reimpports a container from the same source as the original container",
		Args:                  cobra.MinimumNArgs(2),
		RunE:                  CobraRunE(&vars),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := container.ListSources()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
	baseCmd.PersistentFlags().BoolVarP(&vars.fromCache, "fromcache", "c", false, "Reimport the container from the cache, fail if blobs aren't in the cache")
	baseCmd.PersistentFlags().BoolVar(&vars.syncUser, "syncuser", false, "Synchronize UIDs/GIDs from host to container")
	baseCmd.PersistentFlags().BoolVarP(&vars.setBuild, "build", "b", false, "Build container when after pulling")
	baseCmd.PersistentFlags().BoolVar(&vars.ociNoHttps, "ocinohttps", false, "Ignore wrong TLS certificates, superseedes env WAREWULF_OCI_NOHTTPS")
	baseCmd.PersistentFlags().StringVar(&vars.ociUsername, "ociusername", "", "Set username for the access to the registry, superseedes env WAREWULF_OCI_USERNAME")
	baseCmd.PersistentFlags().StringVar(&vars.ociPassword, "ocipasswd", "", "Set password for the access to the registry, superseedes env WAREWULF_OCI_PASSWORD")
	return baseCmd
}
