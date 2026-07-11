package main

import (
	"log"

	"github.com/omjogani/shape-hill/internal/config"
)

func main() {
	c, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("shapehill api: port=%d log_level=%s", c.Port, c.LogLevel)
}
