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
	app.Usage = "Waits for a given condition before returning"
	app.HideVersion = true

	app.Action = shellAction

	app.Flags = []cli.Flag{
		matchFlag,
		exitCodeFlag,
		timeoutFlag,
		intervalFlag,
		verboseFlag,
		failFlag,
	}

	app.Commands = []cli.Command{
		shellCommand,
		portCommand,
		curlCommand,
	}

	return app
}

func main() {
	app().Run(os.Args)
}
