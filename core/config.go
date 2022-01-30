package nodemuxcore

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

	scheme string
}

func (self StoreConfig) Scheme() string {
	return self.scheme
}

type NodemuxConfig struct {
	Version   string                    `yaml:"version,omitempty"`
	Store     StoreConfig               `yaml:"store,omitempty"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints"`
}

// methods
func NewConfig() *NodemuxConfig {
	cfg := &NodemuxConfig{}
	cfg.validateValues()
	return cfg
}

func ConfigFromFile(yamlPath string) (*NodemuxConfig, error) {
	cfg := NewConfig()
	err := cfg.Load(yamlPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self *NodemuxConfig) validateValues() error {
	if self.Version == "" {
		self.Version = "1.0"
	}
	if self.Store.Url == "" {
		self.Store.Url = "redis://127.0.0.1:6379/0"
	}

	// currently nodemux store uses redis
	u, err := url.Parse(self.Store.Url)
	if err != nil {
		return err
	}

	if u.Scheme != "redis" && u.Scheme != "memory" {
		return errors.New("sync source currently only support redis and memory")
	}
	self.Store.scheme = u.Scheme

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

func (self *NodemuxConfig) Load(yamlPath string) error {
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

func (self *NodemuxConfig) LoadYamldata(yamlData []byte) error {
	err := yaml.Unmarshal(yamlData, self)
	if err != nil {
		return err
	}
	return self.validateValues()
}
