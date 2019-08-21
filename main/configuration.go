package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

//const defaultCfgPATH = ""/etc/exquisite/config.yml""
const defaultCfgPATH = "config.yml"

// Configuration structure
type Configuration struct {
	Debug      bool   `yaml:"debug"`
	ServerPort int    `yaml:"server_port"`
	PgDbURL    string `yaml:"pg_db_url"`
	PgDatabase string `yaml:"pg_database"`
	PgUsername string `yaml:"pg_username"`
	PgPassword string `yaml:"pg_password"`
}

// InitConfiguration method
func (c *Configuration) InitConfiguration(debug bool) Configuration {
	configuration := Configuration{}
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

func (c *Configuration) loadConfiguration() *Configuration {
	yamlFile, err := ioutil.ReadFile(defaultCfgPATH)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func (c *Configuration) writeConfiguration(debug bool) *Configuration {
	var conf Configuration
	conf = c.InitConfiguration(debug)
	data, err := yaml.Marshal(&conf)
	if err != nil {
		log.Fatalf("error: %+v", err)
	}

	if _, err := os.Stat(defaultCfgPATH); os.IsNotExist(err) {
		// write to file
		f, err := os.Create(defaultCfgPATH)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(defaultCfgPATH, data, 0644)
		if err != nil {
			log.Fatal(err)
		}

		f.Close()
	}
	return c
}
