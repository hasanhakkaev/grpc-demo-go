package main

import (
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"github.com/hasanhakkaev/yqapp-demo/internal/server"
	"log"
)

func main() {
	cfg, err := conf.Read()
	if err != nil {
		log.Fatalln("reading config failed", err)
	}
	s, err := server.Setup(*cfg)
	if err != nil {
		log.Fatalln("setup producer failed", err)
	}
	if err := server.Run(s); err != nil {
		log.Fatalln("run producer failed", err)
	}
	log.Println("Closing producer...")
}
