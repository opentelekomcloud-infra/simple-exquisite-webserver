package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getUserDir() string {
	res, _ := os.UserHomeDir()
	return res
}

var defaultUserDir = filepath.Join(getUserDir(), ".too-simple")
var defaultCfgPATH = filepath.Join(defaultUserDir, "config.yml")

type PostgresConfig struct {
	DbURL    string `yaml:"db_url"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Configuration file structure
type Configuration struct {
	Debug      bool            `yaml:"debug"`
	ServerPort int             `yaml:"server_port"`
	Postgres   *PostgresConfig `yaml:"postgres,omitempty"`
}

// LoadConfiguration load configuration from given path
func LoadConfiguration(path string) (*Configuration, error) {
	if path == "" {
		path = defaultCfgPATH
	}
	yamlFile, err := ioutil.ReadFile(path)
	cfg := Configuration{ServerPort: 6666}
	if err == nil {
		err = yaml.Unmarshal(yamlFile, &cfg)
	}
	return &cfg, err
}

func createNewConfigFile(path string, data *[]byte) error {
	_ = os.MkdirAll(filepath.Dir(path), 0744)
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
