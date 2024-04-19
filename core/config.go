package nodemuxcore

import (
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"encoding/json"
	"net/url"
	"os"
)

// configs
type EndpointConfig struct {
	Chain         string            `yaml:"chain" json:"chain"`
	Url           string            `yaml:"url" json:"url"`
	StreamingUrl  string            `yaml:"streaming_url,omitempty" json:"streaming_url:omitempty"`
	Headers       map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Weight        int               `yaml:"weight,omitempty" json:"weight,omitempty"`
	SkipMethods   []string          `yaml:"skip_methods,omitempty" json:"skip_methods,omitempty"`
	FetchInterval int               `yaml:"fetch_interval,omitempty" json:"fetch_interval,omitempty"`
	Timeout       int               `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// node specific options
	Options map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

type StoreConfig struct {
	Url string `yaml:"url" json:"url"`
}

func (cfg StoreConfig) Scheme() string {
	u, err := url.Parse(cfg.Url)
	if err != nil {
		panic(err)
	}
	return u.Scheme
}

type NodemuxConfig struct {
	Version   string                    `yaml:"version,omitempty" json:"version,omitempty"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints" json:"endpoints"`
	Stores    map[string]StoreConfig    `yaml:"stores,omitempty" json:"stores,omitempty"`
}

// methods
func NewConfig() *NodemuxConfig {
	cfg := &NodemuxConfig{}
	cfg.validateValues()
	return cfg
}

func ConfigFromFile(configPath string) (*NodemuxConfig, error) {
	cfg := NewConfig()
	err := cfg.Load(configPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *NodemuxConfig) validateValues() error {
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}

	// currently nodemux store uses redis
	for _, store := range cfg.Stores {
		_, err := url.Parse(store.Url)
		if err != nil {
			return err
		}
	}

	for _, epcfg := range cfg.Endpoints {
		if epcfg.Chain == "" {
			return errors.New("empty chain")
		}
		if epcfg.Url == "" {
			return errors.New("empty server url")
		} else if _, err := url.Parse(epcfg.Url); err != nil {
			return errors.Wrap(err, "parse endpoint url")
		}

		if epcfg.StreamingUrl != "" {
			if _, err := url.Parse(epcfg.StreamingUrl); err != nil {
				return errors.Wrap(err, "parse endpoint streaming url")
			}
		}

		for _, skipmtd := range epcfg.SkipMethods {
			if skipmtd == "" {
				return errors.New("empty skip method")
			}
		}

		if epcfg.Options == nil {
			epcfg.Options = make(map[string]interface{})
		}

		if epcfg.FetchInterval <= 0 {
			epcfg.FetchInterval = 1
		}
	}
	return nil
}

func (cfg *NodemuxConfig) Load(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err != nil {
			return err
		}
		return nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	if strings.HasSuffix(configPath, ".json") {
		return cfg.LoadJsondata(data)
	} else {
		return cfg.LoadYamldata(data)
	}
}

func (cfg *NodemuxConfig) LoadYamldata(data []byte) error {
	err := yaml.Unmarshal(data, cfg)
	if err != nil {
		return err
	}
	return cfg.validateValues()
}

func (cfg *NodemuxConfig) LoadJsondata(data []byte) error {
	err := json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}
	return cfg.validateValues()
}
