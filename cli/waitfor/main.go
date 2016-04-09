package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/st3v/waitfor"
)

var (
	waitForConditionWithTimeout = waitfor.ConditionWithTimeout
	exit                        = os.Exit
)

func app() *cli.App {
	app := cli.NewApp()

	app.Name = "waitfor"
	app.Usage = "Waits for a condition before returning"

	app.Commands = []cli.Command{
		portCommand,
	}

	return app
}

func main() {
	app().Run(os.Args)
}
