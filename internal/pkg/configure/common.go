package configure

import (
	"fmt"

	"github.com/pkg/errors"
)

func Configure(serv string, show bool) error {
	if !show {
		fmt.Printf("################################################################################\n")
		fmt.Printf("Configuring: %s\n", serv)
	}

	var err error
	switch serv {
	case "DHCP":
		err = configureDHCP(show)
	case "hosts":
		err = configureHosts(show)
	case "NFS":
		if !show {
			err = configureNFS()
		} else {
			showNFS()
		}
	case "SSH":
		if !show {
			err = configureSSH()
		} else {
			fmt.Printf("'ssh -s' is not yet implemented.\n")
		}
	case "TFTP":
		if !show {
			err = configureTFTP()
		} else {
			fmt.Printf("'tftp -s' is not yet implemented.\n")
		}
	}
	if err != nil {
		return errors.Wrap(err, "Failed to execute configure on "+serv)
	}
	return nil
}
