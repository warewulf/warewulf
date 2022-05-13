package apiconfig

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

// ClientApiConfig contains configuration parameters for an API server.
type ClientApiConfig struct {
	// Server is the hostname or IP address of the server to connect to.
	Server string `yaml:"prefix"`
	// Port is the where the API server listens.
	Port uint32 `yaml:"port"`
}

// ClientConfig is the full client configuration.
type ClientConfig struct {
	ApiConfig ClientApiConfig `yaml:"api"`
	TlsConfig TlsConfig       `yaml:"tls"`
}

// NewClient loads the client config from the given configFilePath.
func NewClient(configFilePath string) (config ClientConfig, err error) {

	log.Printf("Loading api client configuration from: %v\n", configFilePath)

	var fileBytes []byte
	fileBytes, err = ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(fileBytes, &config)
	if err != nil {
		return
	}

	log.Printf("api client config: %#v\n", config)
	return
}
