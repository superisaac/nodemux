package cfg

import (
	"github.com/pkg/errors"
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
	for _, epcfg := range self.Endpoints {
		if epcfg.Chain == "" {
			return errors.New("empty chain")
		}
		if epcfg.Network == "" {
			return errors.New("empty network")
		}
		if epcfg.Url == "" {
			return errors.New("empty server url")
		}
		for _, skipmtd := range epcfg.SkipMethods {
			if skipmtd == "" {
				return errors.New("empty skip method")
			}
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
