package main

import "fmt"

func main() {
	fmt.Printf("Start....\n")
	a := App{}

	var (
		config Configuration
		debug  = true
	)
	config.writeConfiguration("", debug)
	config.loadConfiguration("")
	fmt.Printf("Load config\n")
	a.Initialize(config)
	fmt.Printf("Init app\n")
	a.Run(":6666")
}
