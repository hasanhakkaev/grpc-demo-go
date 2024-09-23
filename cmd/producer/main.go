package main

import (
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"log"
)

func main() {
	cfg, err := conf.Read()
	if err != nil {
		log.Fatalln("reading config failed", err)
	}
	s, err := client.Setup(*cfg)
	if err != nil {
		log.Fatalln("setup consumer failed", err)
	}
	if err := client.Run(s); err != nil {
		log.Fatalln("run consumer failed", err)
	}
	log.Println("Closing consuper...")
}
