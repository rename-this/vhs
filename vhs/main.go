package main

import (
	"log"

	"github.com/gramLabs/vhs/vhs/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
