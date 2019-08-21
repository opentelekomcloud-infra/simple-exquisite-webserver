package main

func main() {
	a := App{}

	var config Configuration
	config.loadConfiguration()
	a.Initialize(config)
	//os.Getenv("APP_DB_USERNAME"),
	//os.Getenv("APP_DB_PASSWORD"),
	//os.Getenv("APP_DB_NAME"))
	a.Run(":6666")
}
