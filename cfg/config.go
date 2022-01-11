package cfg

import (
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
)

func NewConfig() *NodebConfig {
	cfg := &NodebConfig{}
	cfg.validateValues()
	return cfg
}

func ConfigFromFile(yamlPath string) (*NodebConfig, error) {
	cfg := NewConfig()
	err := cfg.Load(yamlPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self *NodebConfig) validateValues() error {
	if self.Version == "" {
		self.Version = "1.0"
	}
	if self.Pubsub.Url != "" {
		u, err := url.Parse(self.Pubsub.Url)
		if err != nil {
			return err
		}
		if u.Scheme != "redis" {
			return errors.New("sync source currently only support redis")
		}
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

func (self *NodebConfig) Load(yamlPath string) error {
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
	return self.LoadYamldata(data)
}

func (self *NodebConfig) LoadYamldata(yamlData []byte) error {
	err := yaml.Unmarshal(yamlData, self)
	if err != nil {
		return err
	}
	return self.validateValues()
}
