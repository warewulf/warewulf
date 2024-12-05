package config

type SSHConf struct {
	KeyTypes []string `yaml:"key types,omitempty" default:"[\"rsa\",\"dsa\",\"ecdsa\",\"ed25519\"]"`
}
