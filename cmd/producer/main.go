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
	producer, err := client.Setup(cfg)
	if err != nil {
		log.Fatalln("setup producer failed", err)
	}
	if err := producer.Run(producer); err != nil {
		log.Fatalln("run producer failed", err)
	}
	log.Println("Closing producer...")
}
