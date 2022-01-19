package server

import (
	//"fmt"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type validConfig interface {
	validateValues() error
}

type TLSConfig struct {
	Certfile string `yaml:"cert"`
	Keyfile  string `yaml:"key"`
}

type BasicAuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type BearerAuthConfig struct {
	Token string `yaml:"token"`
}

type AuthConfig struct {
	Basic  *BasicAuthConfig  `yaml:"basic,omitempty"`
	Bearer *BearerAuthConfig `yaml:"bearer,omitempty"`
}

type MetricsConfig struct {
	Bind string      `yaml:"bind"`
	Auth *AuthConfig `yaml:"auth,omitempty"`
	TLS  *TLSConfig  `yaml:"tls:omitempty"`
}

type EntrypointConfig struct {
	Chain   string      `yaml:"chain"`
	Network string      `yaml:"network"`
	Bind    string      `yaml:version,omitempty`
	Auth    *AuthConfig `yaml:"auth,omitempty"`
	TLS     *TLSConfig  `yaml:"tls:omitempty"`
}

type ServerConfig struct {
	Bind        string              `yaml:version,omitempty`
	TLS         *TLSConfig          `yaml:"tls,omitempty"`
	Metrics     *MetricsConfig      `yaml:"metrics,omitempty"`
	Auth        *AuthConfig         `yaml:"auth,omitempty"`
	Entrypoints []*EntrypointConfig `yaml:"entrypoints,omitempty"`
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
	cfg.Metrics = &MetricsConfig{}
	return cfg
}

func ServerConfigFromFile(yamlPath string) (*ServerConfig, error) {
	cfg := NewServerConfig()
	err := cfg.Load(yamlPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self *ServerConfig) Load(yamlPath string) error {
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

func (self *ServerConfig) LoadYamldata(yamlData []byte) error {
	err := yaml.Unmarshal(yamlData, self)
	if err != nil {
		return err
	}
	return self.validateValues()
}

func (self *ServerConfig) validateValues() error {
	if self.Metrics == nil {
		self.Metrics = &MetricsConfig{}
	}
	children := []validConfig{self.TLS, self.Auth, self.Metrics}
	for _, childcfg := range children {
		if childcfg != nil {
			err := childcfg.validateValues()
			if err != nil {
				return err
			}
		}
	}

	for _, entrycfg := range self.Entrypoints {
		err := entrycfg.validateValues()
		if err != nil {
			return err
		}
	}
	return nil
}

// validators
func (self *TLSConfig) validateValues() error {
	if self == nil {
		return nil
	}
	if self.Certfile == "" {
		return errors.New("ca file is empty")
	}
	if self.Keyfile == "" {
		return errors.New("key file is empty")
	}
	return nil
}

// Auth config
func (self *AuthConfig) validateValues() error {
	if self == nil {
		return nil
	}
	if self.Bearer != nil && self.Bearer.Token == "" {
		return errors.New("bearer token cannot be empty")
	}
	if self.Basic != nil && (self.Basic.Username == "" || self.Basic.Password == "") {
		return errors.New("basic username and password cannot be empty")
	}
	return nil
}

func (self *EntrypointConfig) validateValues() error {
	if self == nil {
		return nil
	}
	if self.Chain == "" {
		return errors.New("entrypoint, chain cannot be empty")
	}
	if self.Network == "" {
		return errors.New("entrypoint, network cannot be empty")
	}
	if self.Bind == "" {
		return errors.New("entrypoint, bind address cannot be empty")
	}
	children := []validConfig{self.TLS, self.Auth}
	for _, cfg := range children {
		if cfg != nil {
			err := cfg.validateValues()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *MetricsConfig) validateValues() error {
	if self == nil {
		return nil
	}
	children := []validConfig{self.TLS, self.Auth}
	for _, cfg := range children {
		if cfg != nil {
			err := cfg.validateValues()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
