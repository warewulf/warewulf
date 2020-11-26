package config

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
)


type Config struct {
	Port            int    `yaml:"warewulfd port", envconfig:"WAREWULFD_PORT"`
	Ipaddr          string `yaml:"warewulfd ipaddr", envconfig:"WAREWULFD_IPADDR"`
	InsecureRuntime bool   `yaml:"insecure runtime"`
	Debug           bool   `yaml:"debug"`
	SysConfDir      string `yaml:"system config dir"`
	LocalStateDir   string `yaml:"local state dir"`
	Editor			string `yaml:"default editor", envconfig:"EDITOR"`
}

var c Config

func init() {
	fd, err := ioutil.ReadFile("/etc/warewulf/warewulf.conf")
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not read config file: %s\n", err)
		os.Exit(255)
	}

	err = yaml.Unmarshal(fd, &c)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not unmarshal config file: %s\n", err)
		os.Exit(255)
	}

	err = envconfig.Process("", &c)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not obtain environment configuration: %s\n", err)
		os.Exit(255)
	}

	if c.Ipaddr == "" {
		fmt.Printf("ERROR: 'warewulf ipaddr' has not been set in /etc/warewulf/warewulf.conf\n")
	}

	if c.SysConfDir == "" {
		c.SysConfDir = "/etc/warewulf"
	}
	if c.LocalStateDir == "" {
		c.LocalStateDir = "/var/warewulf"
	}
	if c.Editor == "" {
		c.Editor = "vi"
	}

	util.ValidateOrDie("warewulfd ipaddr", c.Ipaddr, "^[0-9]+.[0-9]+.[0-9]+.[0-9]+$")
	util.ValidateOrDie("system config dir", c.SysConfDir, "^[a-zA-Z0-9-._:/]+$")
	util.ValidateOrDie("local state dir", c.LocalStateDir, "^[a-zA-Z0-9-._:/]+$")
	util.ValidateOrDie("default editor", c.LocalStateDir, "^[a-zA-Z0-9-._:/]+$")

}

func New() (Config) {
	return c
}

func (self *Config) NodeConfig() string {
	return fmt.Sprintf("%s/nodes.conf", self.LocalStateDir)
}

func (self *Config) OverlayDir() string {
	return fmt.Sprintf("%s/overlays/", self.LocalStateDir)
}

func (self *Config) SystemOverlayDir() string {
	return path.Join(self.OverlayDir(), "/system")
}

func (self *Config) RuntimeOverlayDir() string {
	return path.Join(self.OverlayDir(), "/runtime")
}

func (self *Config) VnfsImageParentDir() string {
	return fmt.Sprintf("%s/provision/vnfs/", self.LocalStateDir)
}

func (self *Config) VnfsChrootParentDir() string {
	return fmt.Sprintf("%s/chroot/", self.LocalStateDir)
}

func (self *Config) KernelParentDir() string {
	return fmt.Sprintf("%s/provision/kernel/", self.LocalStateDir)
}

func (self *Config) SystemOverlaySource(overlayName string) string {
	if overlayName == "" {
		wwlog.Printf(wwlog.ERROR, "System overlay name is not defined\n")
		return ""
	}

	if util.TaintCheck(overlayName, "^[a-zA-Z0-9-._]+$") == false {
		wwlog.Printf(wwlog.ERROR, "System overlay name contains illegal characters: %s\n", overlayName)
		return ""
	}

	return path.Join(self.SystemOverlayDir(), overlayName)
}


func (self *Config) RuntimeOverlaySource(overlayName string) string {
	if overlayName == "" {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name is not defined\n")
		return ""
	}

	if util.TaintCheck(overlayName, "^[a-zA-Z0-9-._]+$") == false {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", overlayName)
		return ""
	}

	return path.Join(self.RuntimeOverlayDir(), overlayName)
}

func (self *Config) KernelImage(kernelVersion string) string {
	if kernelVersion == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Version is not defined\n")
		return ""
	}

	if util.TaintCheck(kernelVersion, "^[a-zA-Z0-9-._]+$") == false {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelVersion)
		return ""
	}

	return path.Join(self.KernelParentDir(), kernelVersion, "vmlinuz")
}

func (self *Config) KmodsImage(kernelVersion string) string {
	if kernelVersion == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Version is not defined\n")
		return ""
	}

	if util.TaintCheck(kernelVersion, "^[a-zA-Z0-9-._]+$") == false {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelVersion)
		return ""
	}

	return path.Join(self.KernelParentDir(), kernelVersion, "kmods.img")
}

func (self *Config) SystemOverlayImage(nodeName string) string {
	if nodeName == "" {
		wwlog.Printf(wwlog.ERROR, "Node name is not defined\n")
		return ""
	}

	if util.TaintCheck(nodeName, "^[a-zA-Z0-9-._:]+$") == false {
		wwlog.Printf(wwlog.ERROR, "System overlay name contains illegal characters: %s\n", nodeName)
		return ""
	}

	return fmt.Sprintf("%s/provision/overlays/system/%s.img", self.LocalStateDir, nodeName)
}

func (self *Config) RuntimeOverlayImage(nodeName string) string {
	if nodeName == "" {
		wwlog.Printf(wwlog.ERROR, "Node name is not defined\n")
		return ""
	}

	if util.TaintCheck(nodeName, "^[a-zA-Z0-9-._:]+$") == false {
		wwlog.Printf(wwlog.ERROR, "System overlay name contains illegal characters: %s\n", nodeName)
		return ""
	}

	return fmt.Sprintf("%s/provision/overlays/runtime/%s.img", self.LocalStateDir, nodeName)
}

func (self *Config) VnfsImageDir(uri string) string {
	if uri == "" {
		wwlog.Printf(wwlog.ERROR, "VNFS URI is not defined\n")
		return ""
	}

	if util.TaintCheck(uri, "^[a-zA-Z0-9-._:]+$") == false {
		wwlog.Printf(wwlog.ERROR, "VNFS name contains illegal characters: %s\n", uri)
		return ""
	}

	return path.Join(self.VnfsImageParentDir(), uri)
}

func (self *Config) VnfsImage(uri string) string {
	return path.Join(self.VnfsImageDir(uri), "image")
}

func (self *Config) VnfsChroot(uri string) string {
	if uri == "" {
		wwlog.Printf(wwlog.ERROR, "VNFS name is not defined\n")
		return ""
	}

	if util.TaintCheck(uri, "^[a-zA-Z0-9-._:]+$") == false {
		wwlog.Printf(wwlog.ERROR, "VNFS name contains illegal characters: %s\n", uri)
		return ""
	}

	return path.Join(self.VnfsChrootParentDir(), uri)
}
