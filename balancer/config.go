package balancer

import (
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
)

// configs
type EndpointConfig struct {
	Chain       string            `yaml:"chain"`
	Network     string            `yaml:"network"`
	Url         string            `yaml:"url"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	SkipMethods []string          `yaml:"skip_methods,omitempty"`
	//HeightPadding int               `yaml:"height_padding,omitempty"`
}

type StoreConfig struct {
	Url string `yaml:"url"`
}

type NodepoolConfig struct {
	Version   string                    `yaml:"version,omitempty"`
	Store     StoreConfig               `yaml:"store,omitempty"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints"`
}

// methods

func NewConfig() *NodepoolConfig {
	cfg := &NodepoolConfig{}
	cfg.validateValues()
	return cfg
}

func ConfigFromFile(yamlPath string) (*NodepoolConfig, error) {
	cfg := NewConfig()
	err := cfg.Load(yamlPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self *NodepoolConfig) validateValues() error {
	if self.Version == "" {
		self.Version = "1.0"
	}
	if self.Store.Url != "" {
		u, err := url.Parse(self.Store.Url)
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

func (self *NodepoolConfig) Load(yamlPath string) error {
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

func (self *NodepoolConfig) LoadYamldata(yamlData []byte) error {
	err := yaml.Unmarshal(yamlData, self)
	if err != nil {
		return err
	}
	return self.validateValues()
}
