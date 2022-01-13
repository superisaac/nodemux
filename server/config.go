package server

import (
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type CertConfig struct {
	CAfile  string `yaml:"ca"`
	Keyfile string `yaml:"key"`
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
	Auth *AuthConfig `yaml:"auth,omitempty"`
}

type EntrypointConfig struct {
	Chain   string `yaml:"chain"`
	Network string `yaml:"network"`
	Bind    string `yaml:version,omitempty`
}

type ServerConfig struct {
	Bind        string             `yaml:version,omitempty`
	Cert        CertConfig         `yaml:"cert,omitempty"`
	Metrics     MetricsConfig      `yaml:"metrics,omitempty"`
	Auth        *AuthConfig        `yaml:"auth,omitempty"`
	Entrypoints []EntrypointConfig `yaml:"entrypoints,omitempty"`
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
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
	for _, entrycfg := range self.Entrypoints {
		if entrycfg.Chain == "" || entrycfg.Network == "" || entrycfg.Bind == "" {
			return errors.New("fields of entrypoint cannot be empty")
		}
	}
	return nil
}

func (self ServerConfig) CertAvailable() bool {
	return self.Cert.CAfile != "" && self.Cert.Keyfile != ""
}

// Auth config
func (self AuthConfig) Available() bool {
	if self.Bearer != nil && self.Bearer.Token != "" {
		return true
	}
	if self.Basic != nil && self.Basic.Username != "" && self.Basic.Password != "" {
		return true
	}
	return false
}
