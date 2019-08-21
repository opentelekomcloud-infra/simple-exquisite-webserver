package main

import (
	"fmt"
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

func (c *Configuration) writeConfiguration() *Configuration {
	data, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatalf("error: %+v", err)
	}
	fmt.Printf("%+v\n", data)

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

	return c
}
