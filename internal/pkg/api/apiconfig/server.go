package apiconfig

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

// ServerApiConfig contains configuration parameters for an API server.
type ServerApiConfig struct {
	// Version contains the full version of the API server, eg: 1.0.0.
	Version string `yaml:"version"`
	// Prefix contains the version url prefix for the API server, eg: v1.
	Prefix string `yaml:"prefix"`
	// Port is the where the API server listens.
	Port uint32 `yaml:"port"`
}

// ServerConfig is the full server configuration.
type ServerConfig struct {
	ApiConfig ServerApiConfig `yaml:"api"`
	TlsConfig TlsConfig       `yaml:"tls"`
}

// NewServer loads the server config from the given configFilePath.
func NewServer(configFilePath string) (config ServerConfig, err error) {

	log.Printf("Loading api server configuration from: %v\n", configFilePath)

	var fileBytes []byte
	fileBytes, err = ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(fileBytes, &config)
	if err != nil {
		return
	}

	log.Printf("api server config: %#v\n", config)
	return
}
