package configure

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

func SSH() error {
	if os.Getuid() == 0 {
		fmt.Printf("Updating system keys\n")

		wwkeydir := path.Join(buildconfig.SYSCONFDIR(), "warewulf/keys") + "/"

		err := os.MkdirAll(path.Join(buildconfig.SYSCONFDIR(), "warewulf/keys"), 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not create base directory: %s\n", err)
			os.Exit(1)
		}

		for _, k := range [4]string{"rsa", "dsa", "ecdsa", "ed25519"} {
			keytype := "ssh_host_" + k + "_key"
			if !util.IsFile(path.Join(wwkeydir, keytype)) {
				fmt.Printf("Setting up key: %s\n", keytype)
				wwlog.Printf(wwlog.DEBUG, "Creating new %s key\n", keytype)
				err = util.ExecInteractive("ssh-keygen", "-q", "-t", k, "-f", path.Join(wwkeydir, keytype), "-C", "", "-N", "")
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Failed to exec ssh-keygen: %s\n", err)
					return errors.Wrap(err, "failed to exec ssh-keygen command")
				}
			} else {
				fmt.Printf("Skipping, key already exists: %s\n", keytype)
			}
		}
	} else {
		fmt.Printf("Updating user's keys\n")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not obtain the user's home directory: %s\n", err)
		os.Exit(1)
	}

	authorizedKeys := path.Join(homeDir, "/.ssh/authorized_keys")
	rsaPriv := path.Join(homeDir, "/.ssh/id_rsa")
	rsaPub := path.Join(homeDir, "/.ssh/id_rsa.pub")

	if !util.IsFile(authorizedKeys) {
		fmt.Printf("Setting up: %s\n", authorizedKeys)
		err = util.ExecInteractive("ssh-keygen", "-q", "-t", "rsa", "-f", rsaPriv, "-C", "", "-N", "")
		if err != nil {
			return errors.Wrap(err, "failed to exec ssh-keygen command")
		}
		err := util.CopyFile(rsaPub, authorizedKeys)
		if err != nil {
			return errors.Wrap(err, "failed to copy keys")
		}
	} else {
		fmt.Printf("Skipping, authorized_keys already exists: %s\n", authorizedKeys)
	}

	return nil
}
