package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/superisaac/jsoff/net"
	yaml "gopkg.in/yaml.v2"
)

// type TLSConfig struct {
// 	Certfile string `yaml:"cert"`
// 	Keyfile  string `yaml:"key"`
// }

type MetricsConfig struct {
	Bind string
	Auth *jsoffnet.AuthConfig `yaml:"auth,omitempty" json:"auth,omitempty"`
	TLS  *jsoffnet.TLSConfig  `yaml:"tls:omitempty" json:"tls:omitempty"`
}

type AdminConfig struct {
	Auth *jsoffnet.AuthConfig `yaml:"auth,omitempty" json:"auth,omitempty"`
}

type EntrypointConfig struct {
	Account string
	Chain   string
	Bind    string
	Auth    *jsoffnet.AuthConfig `yaml:"auth,omitempty" json:"auth,omitempty"`
	TLS     *jsoffnet.TLSConfig  `yaml:"tls,omitempty" json:"tls,omitempty"`
}

type RatelimitConfig struct {
	// requests per IP per hour
	IP int `yaml:"ip" json:"ip"`
	// requests per user per hour
	User int `yaml:"user" json:"user"`
}

type AccountConfig struct {
	Username  string          `yaml:"username" json:"username"`
	Ratelimit RatelimitConfig `yaml:"ratelimit,omitempty" json:"ratelimit,omitempty"`
}

type ServerConfig struct {
	Bind        string                   `yaml:"version,omitempty" json:"version,omitempty"`
	TLS         *jsoffnet.TLSConfig      `yaml:"tls,omitempty" json:"tls,omitempty"`
	Admin       *AdminConfig             `yaml:"admin,omitempty" json:"admin,omitempty"`
	Metrics     *MetricsConfig           `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	Auth        *jsoffnet.AuthConfig     `yaml:"auth,omitempty" json:"auth,omitempty"`
	Entrypoints []EntrypointConfig       `yaml:"entrypoints,omitempty" json:"entrypoints,omitempty"`
	Ratelimit   RatelimitConfig          `yaml:"ratelimit,omitempty" json:"ratelimit,omitempty"`
	Accounts    map[string]AccountConfig `yaml:"accounts,omitempty" json:"accounts,omitempty"`
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
	cfg.Metrics = &MetricsConfig{}
	return cfg
}

type serverConfigKeyType int
var serverConfigKey serverConfigKeyType

func ServerConfigFromContext(ctx context.Context) *ServerConfig {
	if v := ctx.Value(serverConfigKey); v != nil {
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

func (cfg *ServerConfig) AddTo(ctx context.Context) context.Context {
	return context.WithValue(ctx, serverConfigKey, cfg)
}

func (cfg *ServerConfig) Load(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err != nil {
			return err
		}
		return nil
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	if strings.HasSuffix(configPath, ".json") {
		return cfg.LoadJsondata(data)
	} else {
		return cfg.LoadYamldata(data)
	}
}

func (cfg *ServerConfig) LoadYamldata(yamlData []byte) error {
	err := yaml.Unmarshal(yamlData, cfg)
	if err != nil {
		return err
	}
	return cfg.validateValues()
}

func (cfg *ServerConfig) LoadJsondata(yamlData []byte) error {
	err := json.Unmarshal(yamlData, cfg)
	if err != nil {
		return err
	}
	return cfg.validateValues()
}

func (cfg *ServerConfig) validateValues() error {
	if cfg.Metrics == nil {
		cfg.Metrics = &MetricsConfig{}
	}
	if cfg.TLS != nil {
		err := cfg.TLS.ValidateValues()
		if err != nil {
			return err
		}
	}
	if cfg.Auth != nil {
		err := cfg.Auth.ValidateValues()
		if err != nil {
			return err
		}
	}

	if cfg.Metrics != nil {
		err := cfg.Metrics.validateValues()
		if err != nil {
			return err
		}
	}

	for account, acccfg := range cfg.Accounts {
		if strings.Contains(account, "/") || strings.Contains(account, " ") {
			return fmt.Errorf("invalid account name '%s'", account)
		}
		if acccfg.Ratelimit.IP < 0 {
			return fmt.Errorf("acc ip ratelimit < 0, '%s'", account)
		}

		if acccfg.Ratelimit.User < 0 {
			return fmt.Errorf("acc user ratelimit < 0, '%s'", account)
		}

	}

	for _, entrycfg := range cfg.Entrypoints {
		err := entrycfg.validateValues()
		if err != nil {
			return err
		}
	}

	return nil
}

func (cfg *EntrypointConfig) validateValues() error {
	if cfg == nil {
		return nil
	}

	if cfg.Account == "" {
		return errors.New("entrypoint, account cannot be empty")
	}

	if cfg.Chain == "" {
		return errors.New("entrypoint, chain cannot be empty")
	}

	if cfg.Bind == "" {
		return errors.New("entrypoint, bind address cannot be empty")
	}
	if cfg.TLS != nil {
		err := cfg.TLS.ValidateValues()
		if err != nil {
			return err
		}
	}
	if cfg.Auth != nil {
		err := cfg.Auth.ValidateValues()
		if err != nil {
			return err
		}
	}
	return nil
}

func (cfg *MetricsConfig) validateValues() error {
	if cfg == nil {
		return nil
	}

	if cfg.TLS != nil {
		err := cfg.TLS.ValidateValues()
		if err != nil {
			return err
		}
	}
	if cfg.Auth != nil {
		err := cfg.Auth.ValidateValues()
		if err != nil {
			return err
		}
	}
	return nil
}

// Ratelimit Config
func (cfg RatelimitConfig) UserLimit() int {
	if cfg.User <= 0 {
		return 3600
	} else {
		return cfg.User
	}
}

func (cfg RatelimitConfig) IPLimit() int {
	if cfg.IP <= 0 {
		return 3600
	} else {
		return cfg.IP
	}
}
