package main

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"net"
	"net/url"
	"os"
)

func (c *DaemonConnection) init() {
	c.Values = url.Values{}
}

func (c *DaemonConnection) AddInterfaces() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return errors.Wrap(err, "Failed to obtain network interfaces")
	}
	for _, i := range interfaces {
		hwAddr := i.HardwareAddr.String()
		if len(hwAddr) == 0 {
			continue
		}
		c.Values.Add("hwAddr", hwAddr)
	}
	return err
}

func (c *DaemonConnection) AddHostname() error {
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "Failed to get hostname")
	}
	c.Values.Add("name", hostname)
	return err
}

func (c *DaemonConnection) New() error {
	conf, err := warewulfconf.New()
	if err != nil {
		return errors.Wrap(err, "Could not get Warewulf configuration")
	}

	c.updateInterval = conf.Warewulf.UpdateInterval

	if conf.Warewulf.Secure {
		c.TCPAddr.Port = 987
	} else {
		wwlog.Println(wwlog.WARN, "Running from an insecure port")
	}

	// build the URL
	base := fmt.Sprintf("%s:%d", conf.Ipaddr, conf.Warewulf.Port)
	c.URL = url.URL{Scheme: "http", Host: base}
	c.URL.Path += "overlay-runtime"

	wwlog.Printf(wwlog.DEBUG, "baseURL: %s", c.URL.String())

	return err
}
