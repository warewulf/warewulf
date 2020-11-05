package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// THIS IS NOT BEING USED YET AND IS THUS A WORK IN PROGRESS

const ConfigFile = "/etc/warewulf/warewulf.conf"

type Config struct {
		Port   			int    `yaml:"warewulfd port", envconfig:"WAREWULFD_PORT"`
		Ipaddr 			string `yaml:"warewulfd ipaddr", envconfig:"WAREWULFD_IPADDR"`
		InsecureRuntime bool   `yaml:"insecure runtime"`
		Debug  			bool   `yaml:"debug"`
}

func New() (Config, error) {
	var c Config

	fd, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(fd, &c)
	if err != nil {
		return c, err
	}

	err = envconfig.Process("", &c)
	if err != nil {
		return c, err
	}

	if c.Ipaddr == "" {
		fmt.Printf("ERROR: 'warewulf ipaddr' has not been set in %s\n", ConfigFile)
	}

	return c, nil
}
