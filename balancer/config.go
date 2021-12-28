package balancer

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type EndpointConfig struct {
	Chain       string   `yaml:"chain"`
	Network     string   `yaml:"network"`
	Url         string   `yaml:"url"`
	SkipMethods []string `yaml:"skip_methods,omitempty"`
}
type Config struct {
	Version   string                    `yaml:"version,omitempty"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints"`
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.validateValues()
	return cfg
}

func (self *Config) validateValues() error {
	if self.Version == "" {
		self.Version = "1.0"
	}
	return nil
}

func (self *Config) Load(yamlPath string) error {
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		if err != nil {
			return err
		}
		return nil
	}
	data, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return err
	}
	return self.LoadYamlData(data)
}

func (self *Config) LoadYamlData(yamlData []byte) error {
	err := yaml.Unmarshal(yamlData, self)
	if err != nil {
		return err
	}
	return self.validateValues()
}
