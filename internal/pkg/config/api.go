package config

type APIConf struct {
	EnabledP *bool `yaml:"enabled,omitempty" default:"false"`
}

func (conf APIConf) Enabled() bool {
	return BoolP(conf.EnabledP)
}
