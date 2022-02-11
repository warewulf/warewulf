package configure

import (
	"fmt"

	"github.com/pkg/errors"
)

func Configure(s string, v bool) error {
	fmt.Printf("################################################################################\n")
	fmt.Printf("Configuring: %s\n", s)

	var err error
	switch s {
	case "DHCP":
		err = configureDHCP(v)
	case "hosts":
		err = configureHosts(v)
	case "NFS":
		err = configureNFS(v)
	case "SSH":
		err = configureSSH(v)
	case "TFTP":
		err = configureTFTP(v)
	}
	if err != nil {
		return errors.Wrap(err, "Failed to configure "+s)
	}
	return nil
}
