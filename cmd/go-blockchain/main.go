package main

import (
	"os"

	"go-blockchain/internal/cli"
	"go-blockchain/internal/config"
)

func main() {
	app := cli.NewApp(config.Default(), os.Stdout, os.Stderr)
	os.Exit(app.Run(os.Args[1:]))
}
