package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
