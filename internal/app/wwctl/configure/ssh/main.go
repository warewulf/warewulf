package ssh

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	return Configure(SetShow)
}

func Configure(show bool) error {
	if os.Getuid() == 0 {
		fmt.Printf("Updating system keys\n")

		err := os.MkdirAll("/etc/warewulf/keys", 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not create base directory: %s\n", err)
			os.Exit(1)
		}

		if !util.IsFile("/etc/warewulf/keys/ssh_host_rsa_key") {
			fmt.Printf("Setting up key: ssh_host_rsa_key\n")
			err = util.ExecInteractive("ssh-keygen", "-q", "-t", "rsa", "-f", "/etc/warewulf/keys/ssh_host_rsa_key", "-C", "", "-N", "")
			if err != nil {
				return errors.Wrap(err, "failed to exec ssh-keygen command")
			}
		} else {
			fmt.Printf("Skipping, key already exists: ssh_host_rsa_key\n")
		}

		if !util.IsFile("/etc/warewulf/keys/ssh_host_dsa_key") {
			fmt.Printf("Setting up key: ssh_host_dsa_key\n")
			err = util.ExecInteractive("ssh-keygen", "-q", "-t", "dsa", "-f", "/etc/warewulf/keys/ssh_host_dsa_key", "-C", "", "-N", "")
			if err != nil {
				return errors.Wrap(err, "failed to exec ssh-keygen command")
			}
		} else {
			fmt.Printf("Skipping, key already exists: ssh_host_dsa_key\n")
		}

		if !util.IsFile("/etc/warewulf/keys/ssh_host_ecdsa_key") {
			fmt.Printf("Setting up key: ssh_host_ecdsa_key\n")
			err = util.ExecInteractive("ssh-keygen", "-q", "-t", "ecdsa", "-f", "/etc/warewulf/keys/ssh_host_ecdsa_key", "-C", "", "-N", "")
			if err != nil {
				return errors.Wrap(err, "failed to exec ssh-keygen command")
			}
		} else {
			fmt.Printf("Skipping, key already exists: ssh_host_ecdsa_key\n")
		}

		if !util.IsFile("/etc/warewulf/keys/ssh_host_ed25519_key") {
			fmt.Printf("Setting up key: ssh_host_ed25519_key\n")
			err = util.ExecInteractive("ssh-keygen", "-q", "-t", "ed25519", "-f", "/etc/warewulf/keys/ssh_host_ed25519_key", "-C", "", "-N", "")
			if err != nil {
				return errors.Wrap(err, "failed to exec ssh-keygen command")
			}
		} else {
			fmt.Printf("Skipping, key already exists: ssh_host_ed25519_key\n")
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
