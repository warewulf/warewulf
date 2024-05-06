package config

type SSHConf struct {
	KeyTypes []string `yaml:"key types" default:"[\"rsa\",\"dsa\",\"ecdsa\",\"ed25519\"]"`
}
