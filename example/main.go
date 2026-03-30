package main

import (
	"log"
	"time"

	"github.com/crit/bind"
)

type Config struct {
	Name     string    `flag:"name"`
	Age      int       `flag:"age"`
	Birthday time.Time `flag:"birthday"`
}

type BadConfig struct {
	Foo string `flag:"foo"`
}

// go run main.go -- main.go --name=John --age=23 --birthday=2006-01-06
func main() {
	var cfg Config
	bind.RegisterFlags(cfg)

	err := bind.Flag(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = bind.Flag(&BadConfig{})
	if err != nil {
		log.Println(err.Error())
	}

	log.Println(cfg)
}
