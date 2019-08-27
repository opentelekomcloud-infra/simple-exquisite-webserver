package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// initConfiguration method for writeConfiguration func
func (c *Configuration) initConfiguration(debug bool) *Configuration {
	configuration := new(Configuration)
	if debug {
		configuration.Debug = true
	} else {
		configuration.Debug = false
		configuration.PgDatabase = "entities"
		configuration.PgDbURL = "localhost:9999"
		configuration.PgUsername = "entities"
		configuration.PgPassword = ""
		configuration.ServerPort = 5054
	}
	return configuration
}

// LoadConfiguration load configuration from config.yml
func (c *Configuration) LoadConfiguration(path string) *Configuration {
	if path == "" {
		path = defaultCfgPATH
	}
	yamlFile, err := ioutil.ReadFile(path)
	check(err)
	err = yaml.Unmarshal(yamlFile, c)
	check(err)
	return c
}

// WriteConfiguration write config.yml if it not exist with debug or not mode
func (c *Configuration) WriteConfiguration(path string, debug bool) *Configuration {
	if path == "" {
		path = defaultCfgPATH
	}
	var conf = c.initConfiguration(debug)
	data, err := yaml.Marshal(&conf)
	check(err)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// write to file
		f, err := os.Create(filepath.Join(filepath.Dir(path), filepath.Base(path)))
		if err != nil {
			check(err)
		}

		err = ioutil.WriteFile(path, data, 0644)
		check(err)

		f.Close()
	}
	return c
}
