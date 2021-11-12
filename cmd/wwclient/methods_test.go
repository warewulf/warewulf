package main

import (
	"net"
	"net/url"
	"testing"
)

func TestDaemonConnection_AddHostname(t *testing.T) {
	type fields struct {
		URL            url.URL
		TCPAddr        net.TCPAddr
		updateInterval int
		Values         url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"t1", fields{url.URL{Scheme: "http", Host: "test"}, net.TCPAddr{}, 10, url.Values{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DaemonConnection{
				URL:            tt.fields.URL,
				TCPAddr:        tt.fields.TCPAddr,
				updateInterval: tt.fields.updateInterval,
				Values:         tt.fields.Values,
			}
			if err := c.AddHostname(); (err != nil) != tt.wantErr {
				t.Errorf("AddHostname() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDaemonConnection_AddInterfaces(t *testing.T) {
	type fields struct {
		URL            url.URL
		TCPAddr        net.TCPAddr
		updateInterval int
		Values         url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"t1", fields{url.URL{Scheme: "http", Host: "test"}, net.TCPAddr{}, 10, url.Values{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DaemonConnection{
				URL:            tt.fields.URL,
				TCPAddr:        tt.fields.TCPAddr,
				updateInterval: tt.fields.updateInterval,
				Values:         tt.fields.Values,
			}
			if err := c.AddInterfaces(); (err != nil) != tt.wantErr {
				t.Errorf("AddInterfaces() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDaemonConnection_New(t *testing.T) {
	type fields struct {
		URL            url.URL
		TCPAddr        net.TCPAddr
		updateInterval int
		Values         url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"t1", fields{url.URL{}, net.TCPAddr{}, 10, url.Values{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DaemonConnection{
				URL:            tt.fields.URL,
				TCPAddr:        tt.fields.TCPAddr,
				updateInterval: tt.fields.updateInterval,
				Values:         tt.fields.Values,
			}
			if err := c.New(); (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
