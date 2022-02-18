package server

import (
	"context"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz/http"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// type TLSConfig struct {
// 	Certfile string `yaml:"cert"`
// 	Keyfile  string `yaml:"key"`
// }

type MetricsConfig struct {
	Bind string
	Auth *jsonzhttp.AuthConfig `yaml:"auth,omitempty"`
	TLS  *jsonzhttp.TLSConfig  `yaml:"tls:omitempty"`
}

type EntrypointConfig struct {
	Chain   string
	Network string
	Bind    string
	Auth    *jsonzhttp.AuthConfig `yaml:"auth,omitempty"`
	TLS     *jsonzhttp.TLSConfig  `yaml:"tls,omitempty"`
}

type RatelimitConfig struct {
	IP   int `yaml:"ip"`
	User int `yaml:"user"`
}

type ServerConfig struct {
	Bind        string                `yaml:"version,omitempty"`
	TLS         *jsonzhttp.TLSConfig  `yaml:"tls,omitempty"`
	Metrics     *MetricsConfig        `yaml:"metrics,omitempty"`
	Auth        *jsonzhttp.AuthConfig `yaml:"auth,omitempty"`
	Entrypoints []*EntrypointConfig   `yaml:"entrypoints,omitempty"`
	Ratelimit   RatelimitConfig       `yaml:"ratelimit,omitempty"`
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
	cfg.Metrics = &MetricsConfig{}
	return cfg
}

func ServerConfigFromContext(ctx context.Context) *ServerConfig {
	if v := ctx.Value("serverConfig"); v != nil {
		if serverCfg, ok := v.(*ServerConfig); ok {
			return serverCfg
		}
		panic("context value serverConfig is not a serverConfig instance")
	}
	panic("context does not have serverConfig")
}

func ServerConfigFromFile(yamlPath string) (*ServerConfig, error) {
	cfg := NewServerConfig()
	err := cfg.Load(yamlPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self *ServerConfig) AddTo(ctx context.Context) context.Context {
	return context.WithValue(ctx, "serverConfig", self)
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
	if self.TLS != nil {
		err := self.TLS.ValidateValues()
		if err != nil {
			return err
		}
	}
	if self.Auth != nil {
		err := self.Auth.ValidateValues()
		if err != nil {
			return err
		}
	}

	if self.Metrics != nil {
		err := self.Metrics.validateValues()
		if err != nil {
			return err
		}
	}

	for _, entrycfg := range self.Entrypoints {
		err := entrycfg.validateValues()
		if err != nil {
			return err
		}
	}

	if self.Ratelimit.IP <= 0 {
		self.Ratelimit.IP = 3600
	}

	if self.Ratelimit.User <= 0 {
		self.Ratelimit.User = 3600
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
	if self.TLS != nil {
		err := self.TLS.ValidateValues()
		if err != nil {
			return err
		}
	}
	if self.Auth != nil {
		err := self.Auth.ValidateValues()
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *MetricsConfig) validateValues() error {
	if self == nil {
		return nil
	}

	if self.TLS != nil {
		err := self.TLS.ValidateValues()
		if err != nil {
			return err
		}
	}
	if self.Auth != nil {
		err := self.Auth.ValidateValues()
		if err != nil {
			return err
		}
	}
	return nil
}
