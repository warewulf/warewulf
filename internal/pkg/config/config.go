package config

import (
	"gopkg.in/yaml.v2"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
)

type Config struct {
	Port 		int 	`yaml:"port", envconfig:"WAREWULFD_PORT"`
	Ipaddr 		string 	`yaml:"ipaddr", envconfig:"WAREWULFD_IPADDR"`
	Secure		bool	`yaml:"secure port"`
	Debug		bool 	`yaml:"debug"`
}

func New() (Config, error) {
	var c Config

	fd, err := ioutil.ReadFile("/etc/warewulf/warewulf.conf")
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

	return c, nil
}
