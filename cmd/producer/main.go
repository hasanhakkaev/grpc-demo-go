package main

import (
	"flag"
	"fmt"
	"github.com/hasanhakkaev/yqapp-demo/internal/client"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"log"
	"os"
)

var (
	Version   string
	BuildTime string
)

func main() {
	// Define a command-line flag for checking the version
	versionFlag := flag.Bool("version", false, "Print the version of the service")

	// Parse command-line flags
	flag.Parse()

	// If the -version flag is passed, print the version and exit
	if *versionFlag {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		os.Exit(0)
	}

	cfg, err := conf.Read()
	if err != nil {
		log.Fatalln("reading config failed", err)
	}
	c, err := client.Setup(*cfg)
	if err != nil {
		log.Fatalln("setup producer failed", err)
	}
	if err = client.Run(c); err != nil {
		//log.Fatalln("run producer failed", err)
	}
	log.Println("Closing producer...")
}
