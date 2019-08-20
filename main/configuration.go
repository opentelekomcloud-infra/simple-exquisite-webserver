package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

const defaultCfgPATH = "/etc/too-simple/config.yml"

// Configuration structure
type Configuration struct {
	debug bool `yaml:"debug"`

	serverPort int `yaml:"server_port"`

	pgDbURL    string `yaml:"localhost:9999"`
	pgDatabase string `yaml:"entities"`
	pgUsername string `yaml:"entities"`
	pgPassword string `yaml:""`
}

// InitConfiguration create new instance of configuration
func InitConfiguration(debug bool) Configuration {
	configuration := Configuration{}
	if debug == true {
		configuration.debug = debug
		configuration.pgDatabase = "entities"
		configuration.pgUsername = "entities"
	}
	configuration.debug = debug
	configuration.pgDatabase = "entities"
	configuration.pgDbURL = "localhost:9999"
	configuration.pgUsername = "entities"
	configuration.pgPassword = ""
	configuration.serverPort = 5054
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

func (c *Configuration) writeConfiguration() *Configuration {
	data, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", data)
	return c
}

//func main() {
//	var c Configuration
//	c.loadConfiguration()
//
//	fmt.Println(c)
//}

// func write_configuration (path=DEFAULT_CFG_PATH) {

//}

//func load_configuration (path=DEFAULT_CFG_PATH) {

//}
