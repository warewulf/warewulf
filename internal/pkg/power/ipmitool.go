package power

import (
	"os"
	"os/exec"
)

type IPMIResult struct {
	err error
	out string
}

type IPMI struct {
	NodeName  string
	HostName  string
	Port      string
	User      string
	Password  string
	AuthType  string
	Interface string
	result    IPMIResult
}

func (ipmi *IPMI) Result() (string, error) {
	return ipmi.result.out, ipmi.result.err
}

func (ipmi *IPMI) Command(ipmiArgs []string) ([]byte, error) {

	var args []string

	if ipmi.Interface == "" {
		ipmi.Interface = "lan"
	}
	if ipmi.Port == "" {
		ipmi.Port = "623"
	}
	args = append(args, "-I", ipmi.Interface, "-H", ipmi.HostName, "-p", ipmi.Port, "-U", ipmi.User, "-P", ipmi.Password)
	args = append(args, ipmiArgs...)
	ipmiCmd := exec.Command("ipmitool", args...)
	return ipmiCmd.CombinedOutput()
}

func (ipmi *IPMI) InteractiveCommand(ipmiArgs []string) error {

	var args []string

	if ipmi.Interface == "" {
		ipmi.Interface = "lan"
	}
	if ipmi.Port == "" {
		ipmi.Port = "623"
	}
	args = append(args, "-I", ipmi.Interface, "-H", ipmi.HostName, "-p", ipmi.Port, "-U", ipmi.User, "-P", ipmi.Password)
	args = append(args, ipmiArgs...)
	ipmiCmd := exec.Command("ipmitool", args...)
	ipmiCmd.Stdout = os.Stdout
	ipmiCmd.Stdin = os.Stdin
	ipmiCmd.Stderr = os.Stderr
	return ipmiCmd.Run()
}

func (ipmi *IPMI) IPMIInteractiveCommand(args ...string) error {
	return ipmi.InteractiveCommand(args)
}

func (ipmi *IPMI) IPMICommand(args ...string) (string, error) {
	ipmiOut, err := ipmi.Command(args)
	ipmi.result.out = string(ipmiOut)
	ipmi.result.err = err
	return ipmi.result.out, ipmi.result.err

}

//func (ipmi IPMI) PowerOn() (string, error) {
//
//	var args []string
//
//	args = append(args, "chassis", "power", "on")
//	ipmiOut, err := ipmi.Command(args)
//	ipmi.result.out = string(ipmiOut)
//	ipmi.result.err = err
//	return ipmi.result.out, ipmi.result.err
//}

func (ipmi *IPMI) PowerOn() (string, error) {
	return ipmi.IPMICommand("chassis", "power", "on")
}

func (ipmi *IPMI) PowerOff() (string, error) {
	return ipmi.IPMICommand("chassis", "power", "off")
}

func (ipmi *IPMI) PowerCycle() (string, error) {
	return ipmi.IPMICommand("chassis", "power", "cycle")
}

func (ipmi *IPMI) PowerReset() (string, error) {
	return ipmi.IPMICommand("chassis", "power", "reset")
}

func (ipmi *IPMI) PowerSoft() (string, error) {
	return ipmi.IPMICommand("chassis", "power", "soft")
}

func (ipmi *IPMI) PowerStatus() (string, error) {
	return ipmi.IPMICommand("chassis", "power", "status")
}

func (ipmi *IPMI) SDRList() (string, error) {
	return ipmi.IPMICommand("sdr", "list")
}

func (ipmi *IPMI) SensorList() (string, error) {
	return ipmi.IPMICommand("sensor", "list")
}

func (ipmi *IPMI) Console() error {
	return ipmi.IPMIInteractiveCommand("sol", "activate")
}
