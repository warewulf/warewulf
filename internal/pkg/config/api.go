package config

import (
	"net"

	"github.com/creasty/defaults"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

type IPNet net.IPNet

func (n IPNet) MarshalText() ([]byte, error) {
	ipnet := n.IPNet()
	return []byte((&ipnet).String()), nil
}

func (n *IPNet) UnmarshalText(text []byte) error {
	_, network, err := net.ParseCIDR(string(text))
	if err != nil {
		return err
	}
	*n = IPNet(*network)
	return nil
}

func (n IPNet) IPNet() net.IPNet {
	return net.IPNet(n)
}

type APIConf struct {
	EnabledP    *bool   `yaml:"enabled,omitempty"         default:"false"`
	AllowedNets []IPNet `yaml:"allowed subnets,omitempty" default:"[\"127.0.0.0/8\", \"::1/128\"]"`
}

func (conf *APIConf) AllowedIPNets() (allowedIPNets []net.IPNet) {
	for _, allowedIPNet := range conf.AllowedNets {
		allowedIPNets = append(allowedIPNets, allowedIPNet.IPNet())
	}
	return allowedIPNets
}

func (conf *APIConf) Unmarshal(unmarshal func(interface{}) error) error {
	if err := defaults.Set(conf); err != nil {
		return err
	}
	return nil
}

func (conf APIConf) Enabled() bool {
	return util.BoolP(conf.EnabledP)
}
