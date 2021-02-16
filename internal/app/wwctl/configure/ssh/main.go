package ssh

import (
	"fmt"
	"os"
	"path"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if SetPersist == false {
		fmt.Println(cmd.Help())
		os.Exit(0)
	}

	if os.Getuid() == 0 {
		fmt.Printf("Updating system keys\n")

		err := os.MkdirAll("/etc/warewulf/keys", 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not create base directory: %s\n", err)
			os.Exit(1)
		}

		if util.IsFile("/etc/warewulf/keys/ssh_host_rsa_key") == false {
			fmt.Printf("Setting up key: ssh_host_rsa_key\n")
			util.ExecInteractive("ssh-keygen", "-q", "-t", "rsa", "-f", "/etc/warewulf/keys/ssh_host_rsa_key", "-C", "", "-N", "")
		} else {
			fmt.Printf("Skipping, key already exists: ssh_host_rsa_key\n")
		}

		if util.IsFile("/etc/warewulf/keys/ssh_host_dsa_key") == false {
			fmt.Printf("Setting up key: ssh_host_dsa_key\n")
			util.ExecInteractive("ssh-keygen", "-q", "-t", "dsa", "-f", "/etc/warewulf/keys/ssh_host_dsa_key", "-C", "", "-N", "")
		} else {
			fmt.Printf("Skipping, key already exists: ssh_host_dsa_key\n")
		}

		if util.IsFile("/etc/warewulf/keys/ssh_host_ecdsa_key") == false {
			fmt.Printf("Setting up key: ssh_host_ecdsa_key\n")
			util.ExecInteractive("ssh-keygen", "-q", "-t", "ecdsa", "-f", "/etc/warewulf/keys/ssh_host_ecdsa_key", "-C", "", "-N", "")
		} else {
			fmt.Printf("Skipping, key already exists: ssh_host_ecdsa_key\n")
		}

		if util.IsFile("/etc/warewulf/keys/ssh_host_ed25519_key") == false {
			fmt.Printf("Setting up key: ssh_host_ed25519_key\n")
			util.ExecInteractive("ssh-keygen", "-q", "-t", "ed25519", "-f", "/etc/warewulf/keys/ssh_host_ed25519_key", "-C", "", "-N", "")
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

	if util.IsFile(authorizedKeys) == false {
		fmt.Printf("Setting up: %s\n", authorizedKeys)
		util.ExecInteractive("ssh-keygen", "-q", "-t", "rsa", "-f", rsaPriv, "-C", "", "-N", "")
		util.CopyFile(rsaPub, authorizedKeys)
	} else {
		fmt.Printf("Skipping, authorized_keys already exists: %s\n", authorizedKeys)
	}

	return nil
}
