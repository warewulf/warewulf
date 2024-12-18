package power

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type IPMIResult struct {
	err error
	out string
}

type IPMI struct {
	node.IpmiConf
	ShowOnly bool
	Cmd      string
	result   IPMIResult
}

func (ipmi *IPMI) Result() (string, error) {
	return ipmi.result.out, ipmi.result.err
}

func (ipmi *IPMI) getStr() (cmdStr string, err error) {
	if ipmi.Template == "" {
		return "", fmt.Errorf("no ipmi/bmc template specified")
	}
	if !strings.HasPrefix(ipmi.Template, "/") {
		conf := warewulfconf.Get()
		ipmi.Template = path.Join(conf.Paths.Datadir, "warewulf/bmc", ipmi.Template)
	}
	fbuf, err := os.ReadFile(ipmi.Template)
	if err != nil {
		return "", fmt.Errorf("couldn't find the template which defines the ipmi/bmc command: %s", err)
	}
	cmdTmpl, err := template.New("bmc command").Parse(string(fbuf))
	if err != nil {
		return "", err
	}
	var tbuffer bytes.Buffer
	err = cmdTmpl.Execute(&tbuffer, *ipmi)
	if err != nil {
		return "", err
	}
	rg := regexp.MustCompile(`(\r\n?|\n){2,}`)
	cmdStr = rg.ReplaceAllString(tbuffer.String(), " ")
	wwlog.Debug("bmc string is: %s", strings.TrimSpace(cmdStr))
	return strings.TrimSpace(cmdStr), nil

}

func (ipmi *IPMI) Command() ([]byte, error) {
	cmdStr, err := ipmi.getStr()
	if err != nil {
		return []byte{}, err
	}
	if ipmi.ShowOnly {
		return []byte(cmdStr), nil
	}
	ipmiCmd := exec.Command(cmdStr)
	return ipmiCmd.CombinedOutput()
}

func (ipmi *IPMI) InteractiveCommand() (err error) {
	cmdStr, err := ipmi.getStr()
	if err != nil {
		return err
	}
	ipmiCmd := exec.Command(cmdStr)
	ipmiCmd.Stdout = os.Stdout
	ipmiCmd.Stdin = os.Stdin
	ipmiCmd.Stderr = os.Stderr
	return ipmiCmd.Run()
}

func (ipmi *IPMI) IPMIInteractiveCommand(cmd string) error {
	ipmi.Cmd = cmd
	return ipmi.InteractiveCommand()
}

func (ipmi *IPMI) IPMICommand(cmd string) (string, error) {
	ipmi.Cmd = cmd
	ipmiOut, err := ipmi.Command()
	ipmi.result.out = strings.TrimSpace(string(ipmiOut))
	ipmi.result.err = err
	return ipmi.result.out, ipmi.result.err

}

/*
Just define meta commands here, implementation is in the template
*/

func (ipmi *IPMI) PowerOn() (string, error) {
	return ipmi.IPMICommand("PowerOn")
}

func (ipmi *IPMI) PowerOff() (string, error) {
	return ipmi.IPMICommand("PowerOff")
}

func (ipmi *IPMI) PowerCycle() (string, error) {
	return ipmi.IPMICommand("PowerCycle")
}

func (ipmi *IPMI) PowerReset() (string, error) {
	return ipmi.IPMICommand("PowerReset")
}

func (ipmi *IPMI) PowerSoft() (string, error) {
	return ipmi.IPMICommand("PowerSoft")
}

func (ipmi *IPMI) PowerStatus() (string, error) {
	return ipmi.IPMICommand("PowerStatus")
}

func (ipmi *IPMI) SDRList() (string, error) {
	return ipmi.IPMICommand("SDRList")
}

func (ipmi *IPMI) SensorList() (string, error) {
	return ipmi.IPMICommand("SensorList")
}

func (ipmi *IPMI) Console() error {
	return ipmi.IPMIInteractiveCommand("Console")
}
