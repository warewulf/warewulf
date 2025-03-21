package bmc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type Result struct {
	err error
	out string
}

type TemplateStruct struct {
	node.IpmiConf
	ShowOnly bool
	Cmd      string
	result   Result
}

func (tstruct *TemplateStruct) Result() (string, error) {
	return tstruct.result.out, tstruct.result.err
}

func (tstruct *TemplateStruct) getCommand() (cmdStr string, err error) {
	if tstruct.Template == "" {
		return "", fmt.Errorf("no bmc template specified")
	}
	if !strings.HasPrefix(tstruct.Template, "/") {
		conf := warewulfconf.Get()
		tstruct.Template = path.Join(conf.Paths.Datadir, "warewulf/bmc", tstruct.Template)
	}
	fbuf, err := os.ReadFile(tstruct.Template)
	if err != nil {
		return "", fmt.Errorf("couldn't find the template which defines the bmc command: %s", err)
	}
	cmdTmpl, err := template.New("bmc command").Funcs(sprig.TxtFuncMap()).Parse(string(fbuf))
	if err != nil {
		return "", err
	}
	var tbuffer bytes.Buffer
	err = cmdTmpl.Execute(&tbuffer, *tstruct)
	if err != nil {
		return "", err
	}
	cmdStr = strings.TrimSpace(tbuffer.String())
	wwlog.Debug("bmc command: %s", cmdStr)
	return cmdStr, nil

}

func (tstruct *TemplateStruct) runCommand() ([]byte, error) {
	cmdStr, err := tstruct.getCommand()
	if err != nil {
		return []byte{}, err
	}
	if tstruct.ShowOnly {
		return []byte(cmdStr), nil
	}
	return exec.Command("/bin/sh", "-c", cmdStr).CombinedOutput()
}

func (tstruct *TemplateStruct) runInteractiveCommand() (err error) {
	cmdStr, err := tstruct.getCommand()
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (tstruct *TemplateStruct) InteractiveCommand(cmd string) error {
	tstruct.Cmd = cmd
	return tstruct.runInteractiveCommand()
}

func (tstruct *TemplateStruct) Command(cmd string) (string, error) {
	tstruct.Cmd = cmd
	out, err := tstruct.runCommand()
	tstruct.result.out = strings.TrimSpace(string(out))
	tstruct.result.err = err
	return tstruct.result.out, tstruct.result.err
}

/*
Just define meta commands here, implementation is in the template
*/

func (tstruct *TemplateStruct) PowerOn() (string, error) {
	return tstruct.Command("PowerOn")
}

func (tstruct *TemplateStruct) PowerOff() (string, error) {
	return tstruct.Command("PowerOff")
}

func (tstruct *TemplateStruct) PowerCycle() (string, error) {
	return tstruct.Command("PowerCycle")
}

func (tstruct *TemplateStruct) PowerReset() (string, error) {
	return tstruct.Command("PowerReset")
}

func (tstruct *TemplateStruct) PowerSoft() (string, error) {
	return tstruct.Command("PowerSoft")
}

func (tstruct *TemplateStruct) PowerStatus() (string, error) {
	return tstruct.Command("PowerStatus")
}

func (tstruct *TemplateStruct) SDRList() (string, error) {
	return tstruct.Command("SDRList")
}

func (tstruct *TemplateStruct) SensorList() (string, error) {
	return tstruct.Command("SensorList")
}

func (tstruct *TemplateStruct) Console() error {
	return tstruct.InteractiveCommand("Console")
}
