package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

//const defaultCfgPATH = ""/etc/exquisite/config.yml""
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

// loadConfiguration load configuration from config.yml
func (c *Configuration) loadConfiguration(path string) *Configuration {
	if path == "" {
		path = defaultCfgPATH
	}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

// writeConfiguration write config.yml if it not exist with debug or not mode
func (c *Configuration) writeConfiguration(path string, debug bool) *Configuration {
	if path == "" {
		path = defaultCfgPATH
	}
	var conf = c.initConfiguration(debug)
	data, err := yaml.Marshal(&conf)
	if err != nil {
		log.Fatalf("error: %+v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// write to file
		f, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(path, data, 0644)
		if err != nil {
			log.Fatal(err)
		}

		f.Close()
	}
	return c
}
