package main

import (
	"flag"
	"fmt"
	"os"
)

var debug = flag.Bool("debug", false,
	"Enable usage of local database. Taken from config file by default")
var configurationPath = flag.String("config", "", "Set location of Configuration file")

func main() {
	flag.Parse()
	fmt.Printf("Start....\n")
	a := App{}

	path := *configurationPath
	config, err := LoadConfiguration(path)
	//noinspection ALL
	config.Debug = config.Debug || *debug

	if err != nil {
		if os.IsNotExist(err) {
			err := config.WriteConfiguration(path) // write empty config in case no config exists
			if err != nil {
				panic(err)
			}
		}
	}
	fmt.Printf("Load config\n")
	a.Initialize(config)
	fmt.Printf("Init app\n")
	a.Run(fmt.Sprintf(":%v", config.ServerPort))
}
