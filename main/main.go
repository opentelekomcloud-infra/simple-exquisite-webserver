package main

import (
	"flag"
	"fmt"
	"github.com/sevlyar/go-daemon"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

const version = "0.4.1"

func allowedDir(path string) bool {
	fileName := filepath.Join(path, "tmp.tmp")
	_ = os.MkdirAll(path, 0744)
	f, err := os.Create(fileName)
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(fileName)
	return true
}

func selectDir(preferred string, backup string) string {
	if allowedDir(preferred) {
		return preferred
	}
	return backup
}

func termHandler(sig os.Signal) error {
	log.Println("terminating...")

	return daemon.ErrStop
}

func main() {
	debug := flag.Bool("debug", false, "Enable usage of local database. Taken from config file by default")
	showVersion := flag.Bool("showVersion", false, "Print application showVersion")
	configurationPath := flag.String("config", "", "Set location of Configuration file")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	action := "start"
	if flag.NArg() > 0 { // in case there is positional argument
		action = flag.Arg(0)
	}
	daemon.AddCommand(daemon.StringFlag(&action, "stop"), syscall.SIGTERM, termHandler)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	context := &daemon.Context{
		PidFileName: filepath.Join(selectDir("/tmp", defaultUserDir), "too-simple.pid"),
		PidFilePerm: 0644,
		LogFileName: filepath.Join(selectDir("/var/log/too-simple", defaultUserDir), "execution.log"),
		LogFilePerm: 0640,
	}

	if len(daemon.ActiveFlags()) > 0 {
		dProcess, err := context.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		_ = daemon.SendCommands(dProcess)
		return
	}

	a := App{}

	cfgPath := *configurationPath
	config, err := LoadConfiguration(cfgPath)
	//noinspection ALL
	config.Debug = config.Debug || *debug

	if err != nil {
		if os.IsNotExist(err) {
			if err = config.WriteConfiguration(cfgPath); err != nil {
				panic(err)
			}
		}
	}
	log.Print("Load config\n")
	a.Initialize(config)
	log.Print("Init app\n")

	d, err := context.Reborn()
	if err != nil {
		// seems you're running this in windows
		log.Println("Can't start service. Starting in foreground\n", err)
		a.Run(fmt.Sprintf(":%v", config.ServerPort))
	}
	if d != nil { // this is parent process
		return
	}
	defer func() {
		err = context.Release()
		log.Println("Error on closing: ", err)
	}()

	log.Println("Daemon started")

	go a.Run(fmt.Sprintf(":%v", config.ServerPort))

	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}
