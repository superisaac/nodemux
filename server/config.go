package server

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type CertConfig struct {
	CAfile  string `yaml:"ca"`
	Keyfile string `yaml:"key"`
}

type ServerConfig struct {
	Bind string     `yaml:version,omitempty`
	Cert CertConfig `yaml:"cert,omitempty"`
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
	return nil
}

func (self ServerConfig) CertAvailable() bool {
	return self.Cert.CAfile != "" && self.Cert.Keyfile != ""
}
