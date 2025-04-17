package main

import (
	"github.com/awnumar/memguard"
	"github.com/eddie-energy/aiida-helm/installer/internal/app"
)

func main() {
	memguard.CatchInterrupt()
	defer memguard.Purge()

	a := app.App{}
	a.Run(":8010")
}
