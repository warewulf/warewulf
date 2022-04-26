package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/version"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	fmt.Println("VERSION:\t", version.GetVersion())
	fmt.Println("BINDIR:\t", buildconfig.BINDIR())
	fmt.Println("DATADIR:\t", buildconfig.DATADIR())
	fmt.Println("SYSCONFDIR:\t", buildconfig.SYSCONFDIR())
	fmt.Println("LOCALSTATEDIR:\t", buildconfig.LOCALSTATEDIR())
	fmt.Println("SRVDIR:\t", buildconfig.SRVDIR())
	fmt.Println("TFTPDIR:\t", buildconfig.TFTPDIR())
	fmt.Println("SYSTEMDDIR:\t", buildconfig.SYSTEMDDIR())
	fmt.Println("WWOVERLAYDIR:\t", buildconfig.WWOVERLAYDIR())
	fmt.Println("WWCHROOTDIR:\t", buildconfig.WWCHROOTDIR())
	fmt.Println("WWPROVISIONDIR:\t", buildconfig.WWPROVISIONDIR())
	fmt.Println("WWCLIENTDIR:\t", buildconfig.WWCLIENTDIR())

	return nil
}
