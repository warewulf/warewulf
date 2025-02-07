package config

type SSHConf struct {
	KeyTypes []string `yaml:"key types,omitempty" default:"[\"ed25519\",\"ecdsa\",\"rsa\",\"dsa\"]"`
}
