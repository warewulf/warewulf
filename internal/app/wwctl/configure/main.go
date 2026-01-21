package configure

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/configure"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error

	conf := warewulfconf.Get()
	if conf.Autodetected() && conf.InitializedFromFile() {
		if err = conf.PersistToFile(conf.GetWarewulfConf()); err != nil {
			wwlog.Warn("error when persisting auto-detected settings: %s", err)
		}
	}

	if allFunctions {
		keystore := path.Join(conf.Paths.Sysconfdir, "warewulf", "keys")
		keyFile := path.Join(keystore, "warewulf.key")
		certFile := path.Join(keystore, "warewulf.crt")

		if !util.IsFile(keyFile) || !util.IsFile(certFile) {
			err = configure.GenTLSKeys()
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("Keys already exist in %s\n", keystore)
		}

		err = configure.WAREWULFD()
		if err != nil {
			return err
		}

		err = configure.TFTP()
		if err != nil {
			return err
		}

		err = configure.DHCP()
		if err != nil {
			return err
		}

		err = configure.NFS()
		if err != nil {
			return err
		}

		err = configure.SSH(warewulfconf.Get().SSH.KeyTypes...)
		if err != nil {
			return err
		}

		err = configure.Hostfile()
		if err != nil {
			return err
		}

	} else {
		_ = cmd.Help()
		os.Exit(0)
	}

	return nil
}
