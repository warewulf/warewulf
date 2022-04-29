package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/version"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if ListFull {
		fmt.Printf("%s=%s\n", "VERSION", version.GetVersion())
		fmt.Printf("%s=%s\n", "BINDIR", buildconfig.BINDIR())
		fmt.Printf("%s=%s\n", "DATADIR", buildconfig.DATADIR())
		fmt.Printf("%s=%s\n", "SYSCONFDIR", buildconfig.SYSCONFDIR())
		fmt.Printf("%s=%s\n", "LOCALSTATEDIR", buildconfig.LOCALSTATEDIR())
		fmt.Printf("%s=%s\n", "SRVDIR", buildconfig.SRVDIR())
		fmt.Printf("%s=%s\n", "TFTPDIR", buildconfig.TFTPDIR())
		fmt.Printf("%s=%s\n", "SYSTEMDDIR", buildconfig.SYSTEMDDIR())
		fmt.Printf("%s=%s\n", "WWOVERLAYDIR", buildconfig.WWOVERLAYDIR())
		fmt.Printf("%s=%s\n", "WWCHROOTDIR", buildconfig.WWCHROOTDIR())
		fmt.Printf("%s=%s\n", "WWPROVISIONDIR", buildconfig.WWPROVISIONDIR())
		fmt.Printf("%s=%s\n", "BASEVERSION", buildconfig.VERSION())
		fmt.Printf("%s=%s\n", "RELEASE", buildconfig.RELEASE())
		fmt.Printf("%s=%s\n", "WWCLIENTDIR", buildconfig.WWCLIENTDIR())o
	} else {
		fmt.Println(version.GetVersion())
	}
	return nil
}
