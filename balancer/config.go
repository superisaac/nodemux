package balancer

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func NewConfig() *Config {
	cfg := &Config{}
	cfg.validateValues()
	return cfg
}

func ConfigFromFile(yamlPath string) (*Config, error) {
	cfg := NewConfig()
	err := cfg.Load(yamlPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self *Config) validateValues() error {
	if self.Version == "" {
		self.Version = "1.0"
	}
	for _, ep := range self.Endpoints {
		if ep.HeightPadding <= 0 {
			// The default height padding considered safe
			ep.HeightPadding = 2
		}
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
