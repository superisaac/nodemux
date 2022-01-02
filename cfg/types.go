package cfg

import ()

// configs
type EndpointConfig struct {
	Chain         string   `yaml:"chain"`
	Network       string   `yaml:"network"`
	Url           string   `yaml:"url"`
	SkipMethods   []string `yaml:"skip_methods,omitempty"`
	HeightPadding int      `yaml:"height_padding,omitempty"`
}

type Config struct {
	Version   string                    `yaml:"version,omitempty"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints"`
}
