package power

import (
	"os/exec"
)

type IPMI struct {
	HostName string
	User     string
	Password string
	AuthType string
}

func (ipmi IPMI) Command(ipmiArgs []string) ([]byte, error) {

	var args []string

	args = append(args, "-I", "lan", "-H", ipmi.HostName, "-U", ipmi.User, "-P", ipmi.Password)
	args = append(args, ipmiArgs...)
	ipmiCmd := exec.Command("/usr/bin/ipmitool", args...)
	return ipmiCmd.CombinedOutput()
}

func (ipmi IPMI) PowerOn() (string, error) {

	var args []string

	args = append(args, "chassis", "power", "on")
	ipmiOut, err := ipmi.Command(args)
	return string(ipmiOut), err
}

func (ipmi IPMI) PowerOff() (string, error) {

	var args []string

	args = append(args, "chassis", "power", "off")
	ipmiOut, err := ipmi.Command(args)
	return string(ipmiOut), err
}

func (ipmi IPMI) PowerStatus() (string, error) {

	var args []string

	args = append(args, "chassis", "power", "status")
	ipmiOut, err := ipmi.Command(args)
	return string(ipmiOut), err
}
