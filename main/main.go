package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Start....\n")
	a := App{}

	path := ""
	config, err := LoadConfiguration(path)
	//noinspection ALL
	config.Debug = true
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
	a.Run(":6666")
}
