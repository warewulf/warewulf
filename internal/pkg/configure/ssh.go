package configure

import (
	"fmt"
	"os"
	"path"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
Create password less ssh keys in the home of the user who its
calling this function. (root in our case)
*/
func SSH(keyTypes ...string) error {
	if os.Getuid() == 0 {
		fmt.Printf("Updating system keys\n")
		conf := warewulfconf.Get()
		wwkeydir := path.Join(conf.Paths.Sysconfdir, "warewulf/keys") + "/"

		err := os.MkdirAll(path.Join(conf.Paths.Sysconfdir, "warewulf/keys"), 0755)
		if err != nil {
			return fmt.Errorf("could not create base directory: %s", err)
		}

		for _, k := range keyTypes {
			keytype := "ssh_host_" + k + "_key"
			if !util.IsFile(path.Join(wwkeydir, keytype)) {
				fmt.Printf("Setting up key: %s\n", keytype)
				wwlog.Debug("Creating new %s key", keytype)
				err = util.ExecInteractive("ssh-keygen", "-q", "-t", k, "-f", path.Join(wwkeydir, keytype), "-C", "", "-N", "")
				if err != nil {
					wwlog.Error("Failed to exec ssh-keygen: %s", err)
					return fmt.Errorf("failed to exec ssh-keygen command: %w", err)
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
		return fmt.Errorf("could not obtain the user's home directory: %s", err)
	}

	authorizedKeys := path.Join(homeDir, "/.ssh/authorized_keys")

	if !util.IsFile(authorizedKeys) {
		if len(keyTypes) > 0 {
			keyType := keyTypes[0]
			fmt.Printf("Setting up: %s\n", authorizedKeys)
			privKey := path.Join(homeDir, "/.ssh/id_"+keyType)
			pubKey := privKey + ".pub"
			err = util.ExecInteractive("ssh-keygen", "-q", "-t", keyType, "-f", privKey, "-C", "", "-N", "")
			if err != nil {
				return fmt.Errorf("failed to exec ssh-keygen command: %w", err)
			}
			err := util.CopyFile(pubKey, authorizedKeys)
			if err != nil {
				return fmt.Errorf("failed to copy %s to authorized_keys: %w", pubKey, err)
			}
		} else {
			fmt.Printf("Skipping authorized_keys: no key types configured\n")
		}
	} else {
		fmt.Printf("Skipping authorized_keys: already exists: %s\n", authorizedKeys)
	}

	return nil
}
