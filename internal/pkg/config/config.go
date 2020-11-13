package config

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)


type Config struct {
	Port            int    `yaml:"warewulfd port", envconfig:"WAREWULFD_PORT"`
	Ipaddr          string `yaml:"warewulfd ipaddr", envconfig:"WAREWULFD_IPADDR"`
	InsecureRuntime bool   `yaml:"insecure runtime"`
	Debug           bool   `yaml:"debug"`
	SysConfDir      string `yaml:"system config dir"`
	LocalStateDir   string `yaml:"local state dir"`
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
}

func New() (Config) {
	return c
}

func (self *Config) NodeConfig() (string) {
	return fmt.Sprintf("%s/nodes.conf", self.LocalStateDir)
}

func (self *Config) SystemOverlaySource(overlayName string) (string) {
	return fmt.Sprintf("%s/overlays/system/%s", self.LocalStateDir, overlayName)
}

func (self *Config) RuntimeOverlaySource(overlayName string) (string) {
	return fmt.Sprintf("%s/overlays/runtime/%s", self.LocalStateDir, overlayName)
}

func (self *Config) KernelImage(kernelVersion string) (string) {
	return fmt.Sprintf("%s/provision/kernel/vmlinuz-%s", self.LocalStateDir, kernelVersion)
}

func (self *Config) KmodsImage(kernelVersion string) (string) {
	return fmt.Sprintf("%s/provision/kernel/kmods-%s.img", self.LocalStateDir, kernelVersion)
}

func (self *Config) VnfsImage(vnfsNameClean string) (string) {
	return fmt.Sprintf("%s/provision/vnfs/%s.img.gz", self.LocalStateDir, vnfsNameClean)
}

func (self *Config) SystemOverlayImage(nodeName string) (string) {
	return fmt.Sprintf("%s/provision/overlay/system/%s.img", self.LocalStateDir, nodeName)
}

func (self *Config) RuntimeOverlayImage(nodeName string) (string) {
	return fmt.Sprintf("%s/provision/overlay/runtime/%s.img", self.LocalStateDir, nodeName)
}

func (self *Config) VnfsChroot(vnfsNameClean string) (string) {
	return fmt.Sprintf("%s/chroot/%s.img", self.LocalStateDir, vnfsNameClean)
}

