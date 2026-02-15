package main

import (
	"log"

	"agenteur.ai/api/internal/app"
)

func main() {
	app := app.NewApp()
	err := app.Start()
	if err != nil {
		log.Fatal(err)
	}
}
