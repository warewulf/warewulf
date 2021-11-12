package main

import (
	"net"
	"net/url"
)

type DaemonConnection struct {
	URL            url.URL
	TCPAddr        net.TCPAddr
	updateInterval int
	Values         url.Values
}
