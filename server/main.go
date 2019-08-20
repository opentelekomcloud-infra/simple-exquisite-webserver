package main

import (
	"fmt"
	"os"
)

func main() {
	a := App{}

	var c Configuration
	c.writeConfiguration()
	fmt.Println(c)
	//c.loadConfiguration()
	//fmt.Println(c)

	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	a.Run(":6666")
}
