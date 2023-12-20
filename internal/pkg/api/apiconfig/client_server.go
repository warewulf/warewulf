package apiconfig

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// ClientServerConfig is the full client server configuration.
// wwapird is a client of wwapid.
// wwapird serves REST. (WareWulf API Rest Daemon)
type ClientServerConfig struct {
	ClientApiConfig ClientApiConfig `yaml:"clientapi"`
	ServerApiConfig ServerApiConfig `yaml:"serverapi"`
	ClientTlsConfig TlsConfig       `yaml:"clienttls"`
	ServerTlsConfig TlsConfig       `yaml:"servertls"`
}

// NewClientServer loads the client config from the given configFilePath.
func NewClientServer(configFilePath string) (config ClientServerConfig, err error) {

	log.Printf("Loading api client server configuration from: %v\n", configFilePath)

	var fileBytes []byte
	fileBytes, err = os.ReadFile(configFilePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(fileBytes, &config)
	if err != nil {
		return
	}

	log.Printf("api client server config: %#v\n", config)
	return
}
