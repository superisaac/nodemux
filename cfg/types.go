package cfg

import ()

// configs
type EndpointConfig struct {
	Chain       string            `yaml:"chain"`
	Network     string            `yaml:"network"`
	Url         string            `yaml:"url"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	SkipMethods []string          `yaml:"skip_methods,omitempty"`
	//HeightPadding int               `yaml:"height_padding,omitempty"`
}

type PubsubConfig struct {
	Url string `yaml:"url"`
}

type NodebConfig struct {
	Version   string                    `yaml:"version,omitempty"`
	Pubsub    PubsubConfig              `yaml:"pubsub,omitempty"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints"`
}
