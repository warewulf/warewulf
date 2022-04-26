package version

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/version"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	fmt.Printf("%-20s %-18s\n", "VERSION:", version.GetVersion())
	fmt.Printf("%-20s %-18s\n", "BINDIR:", buildconfig.BINDIR())
	fmt.Printf("%-20s %-18s\n", "DATADIR:", buildconfig.DATADIR())
	fmt.Printf("%-20s %-18s\n", "SYSCONFDIR:", buildconfig.SYSCONFDIR())
	fmt.Printf("%-20s %-18s\n", "LOCALSTATEDIR:", buildconfig.LOCALSTATEDIR())
	fmt.Printf("%-20s %-18s\n", "SRVDIR:", buildconfig.SRVDIR())
	fmt.Printf("%-20s %-18s\n", "TFTPDIR:", buildconfig.TFTPDIR())
	fmt.Printf("%-20s %-18s\n", "SYSTEMDDIR:", buildconfig.SYSTEMDDIR())
	fmt.Printf("%-20s %-18s\n", "WWOVERLAYDIR:", buildconfig.WWOVERLAYDIR())
	fmt.Printf("%-20s %-18s\n", "WWCHROOTDIR:", buildconfig.WWCHROOTDIR())
	fmt.Printf("%-20s %-18s\n", "WWPROVISIONDIR:", buildconfig.WWPROVISIONDIR())
	fmt.Printf("%-20s %-18s\n", "BASEVERSION:", buildconfig.VERSION())
	fmt.Printf("%-20s %-18s\n", "RELEASE:", buildconfig.RELEASE())
	fmt.Printf("%-20s %-18s\n", "WWCLIENTDIR:", buildconfig.WWCLIENTDIR())

	return nil
}
