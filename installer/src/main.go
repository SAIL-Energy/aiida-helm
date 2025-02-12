package main

import "github.com/awnumar/memguard"

func main() {
	memguard.CatchInterrupt()
	defer memguard.Purge()

	app := App{}
	app.Run(":8010")
}
