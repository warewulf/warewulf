package imprt

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "import [OPTIONS] SOURCE [NAME]",
		Short:                 "Import an image into Warewulf",
		Aliases:               []string{"pull"},
		Long: `This command will pull and import an image into Warewulf from SOURCE,
optionally renaming it to NAME. The SOURCE must be in a supported URI format. Formats
are:
 * docker://registry.example.org/example:latest
 * docker-daemon://example:latest
 * file://path/to/archive/tar/ball
 * /path/to/archive/tar/ball
 * /path/to/chroot/
Imported images are used to create bootable images.`,
		Example: "wwctl image import docker://ghcr.io/warewulf/warewulf-rockylinux:8 rockylinux-8",
		RunE:    CobraRunE,
		Args:    cobra.RangeArgs(1, 2),
		PreRun: func(cmd *cobra.Command, args []string) {
			if SetForce && SetUpdate {
				wwlog.Warn("Both --force and --update flags are set, will ignore --update flag")
			}
		},
	}
	SetForce    bool
	SetUpdate   bool
	SetBuild    bool
	SyncUser    bool
	OciNoHttps  bool
	OciUsername string
	OciPassword string
	Platform    string
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SetForce, "force", "f", false, "Force overwrite of an existing image")
	baseCmd.PersistentFlags().BoolVarP(&SetUpdate, "update", "u", false, "Update and overwrite an existing image")
	baseCmd.PersistentFlags().BoolVarP(&SetBuild, "build", "b", false, "Build image after pulling")
	baseCmd.PersistentFlags().BoolVar(&SyncUser, "syncuser", false, "Synchronize UIDs/GIDs from host to image")
	baseCmd.PersistentFlags().BoolVar(&OciNoHttps, "nohttps", false, "Ignore wrong TLS certificates, superseedes env WAREWULF_OCI_NOHTTPS")
	baseCmd.PersistentFlags().StringVar(&OciUsername, "username", "", "Set username for the access to the registry, superseedes env WAREWULF_OCI_USERNAME")
	baseCmd.PersistentFlags().StringVar(&OciPassword, "password", "", "Set password for the access to the registry, superseedes env WAREWULF_OCI_PASSWORD")
	baseCmd.PersistentFlags().StringVar(&OciPassword, "passwd", "", "Set password for the access to the registry, superseedes env WAREWULF_OCI_PASSWORD")
	_ = baseCmd.PersistentFlags().MarkHidden("passwd")
	baseCmd.PersistentFlags().StringVar(&Platform, "platform", "", "Set other hardware platform e.g. amd64 or arm64, superseedes env WAREWULF_OCI_PLATFORM")
	baseCmd.PersistentFlags().StringVar(&Platform, "arch", "", "Set other hardware platform, superseedes env WAREWULF_OCI_PLATFORM")
	_ = baseCmd.PersistentFlags().MarkHidden("arch")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
