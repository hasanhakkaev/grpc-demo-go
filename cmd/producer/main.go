package main

import (
	"github.com/hasanhakkaev/yqapp-demo/internal/client"
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	"log"
)

func main() {
	cfg, err := conf.Read()
	if err != nil {
		log.Fatalln("reading config failed", err)
	}
	c, err := client.Setup(*cfg)
	if err != nil {
		log.Fatalln("setup producer failed", err)
	}
	if err := client.Run(c); err != nil {
		log.Fatalln("run producer failed", err)
	}
	log.Println("Closing producer...")
}
