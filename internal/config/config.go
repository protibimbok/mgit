package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Profile struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	SSHKey string `json:"ssh_key"`
}

type Config struct {
	Profiles []Profile `json:"profiles"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mgit", "profiles.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) FindByKey(key string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Key == key {
			return &c.Profiles[i]
		}
	}
	return nil
}

func (c *Config) Add(p Profile) error {
	if c.FindByKey(p.Key) != nil {
		return fmt.Errorf("profile with key %q already exists", p.Key)
	}
	c.Profiles = append(c.Profiles, p)
	return nil
}

func (c *Config) Remove(key string) bool {
	for i, p := range c.Profiles {
		if p.Key == key {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return true
		}
	}
	return false
}
