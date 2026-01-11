package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// CTLog represents a single Certificate Transparency log entry
type CTLog struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Period      string `yaml:"period"`
	Mode        string `yaml:"mode"`
	TreeSize    string `yaml:"treeSize"`
}

// Provider represents a CT log provider with its base URL and logs
type Provider struct {
	BaseURL string  `yaml:"baseUrl"`
	Logs    []CTLog `yaml:"logs"`
}

// Config represents the entire configuration structure
type Config struct {
	Google          Provider `yaml:"google"`
	Cloudflare      Provider `yaml:"cloudflare"`
	Digicert        Provider `yaml:"digicert"`
	LetsEncrypt     Provider `yaml:"letsencrypt"`
	SectigoMammoth  Provider `yaml:"sectigo_mammoth"`
	SectigoElephant Provider `yaml:"sectigo_elephant"`
}

// Global config instance
var cfg *Config

// Load reads the configuration file and populates the global config
func Load(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg = &c
	return nil
}

// Get returns the global config instance
func Get() *Config {
	if cfg == nil {
		panic("config not loaded - call Load() first")
	}
	return cfg
}

// AllLogs returns all CT logs across all providers with their base URLs
func (c *Config) AllLogs() []struct {
	Provider string
	BaseURL  string
	Log      CTLog
} {
	var allLogs []struct {
		Provider string
		BaseURL  string
		Log      CTLog
	}

	providers := []struct {
		Name     string
		Provider Provider
	}{
		{"google", c.Google},
		{"cloudflare", c.Cloudflare},
		{"digicert", c.Digicert},
		{"letsencrypt", c.LetsEncrypt},
		{"sectigo_mammoth", c.SectigoMammoth},
		{"sectigo_elephant", c.SectigoElephant},
	}

	for _, p := range providers {
		for _, log := range p.Provider.Logs {
			allLogs = append(allLogs, struct {
				Provider string
				BaseURL  string
				Log      CTLog
			}{
				Provider: p.Name,
				BaseURL:  p.Provider.BaseURL,
				Log:      log,
			})
		}
	}

	return allLogs
}
