package balancer

import ()

type EndpointConfig struct {
	Chain       string   `yaml:"chain"`
	Network     string   `yaml:"network"`
	Url         string   `yaml:"url"`
	SkipMethods []string `yaml:"skip_methods"`
}
type Config struct {
	Version   string           `yaml:"version"`
	Endpoints []EndpointConfig `yaml:"endpoints"`
}
