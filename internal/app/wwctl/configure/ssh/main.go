package ssh

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

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

	return nil
}
