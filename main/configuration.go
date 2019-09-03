package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const defaultCfgPATH = "config.yml"

// Configuration file structure
type Configuration struct {
	Debug      bool   `yaml:"debug"`
	ServerPort int    `yaml:"server_port"`
	PgDbURL    string `yaml:"pg_db_url"`
	PgDatabase string `yaml:"pg_database"`
	PgUsername string `yaml:"pg_username"`
	PgPassword string `yaml:"pg_password"`
}

// LoadConfiguration load configuration from test_config.yml
func LoadConfiguration(path string) (*Configuration, error) {
	if path == "" {
		path = defaultCfgPATH
	}
	yamlFile, err := ioutil.ReadFile(path)
	cfg := Configuration{}
	if err == nil {
		err = yaml.Unmarshal(yamlFile, &cfg)
	}
	return &cfg, err
}

func createNewConfigFile(path string, data *[]byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.Write(*data)
	_ = f.Close()
	return err
}

// WriteConfiguration write test_config.yml if it not exist with debug or not mode
func (c *Configuration) WriteConfiguration(path string) error {
	if path == "" {
		path = defaultCfgPATH
	}
	data, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// write to file
			return createNewConfigFile(path, &data)
		}
		return err
	}
	return nil
}
