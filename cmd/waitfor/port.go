package main

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/codegangsta/cli"

	"github.com/st3v/waitfor/check"
)

var portCheckProvider = check.Port

var port = func(c *cli.Context) int {
	if !c.Args().Present() {
		cli.ShowCommandHelp(c, "port")
		fmt.Fprintln(c.App.Writer, "must specify port")
		exit(1)
	}

	port, err := strconv.Atoi(c.Args().First())
	if err != nil {
		fmt.Fprintln(c.App.Writer, "invalid port")
		exit(1)
	}

	return port
}

var portCommand = cli.Command{
	Name:  "port",
	Usage: "wait for host to listen on port (or not)",

	HideHelp: true,

	Flags: []cli.Flag{
		closedFlag,
		hostFlag,
		networkFlag,
		timeoutFlag,
		intervalFlag,
		verboseFlag,
	},

	Action: func(c *cli.Context) error {
		host := c.String("host")
		network := c.String("network")
		timeout := c.Duration("timeout")
		interval := c.Duration("interval")

		port := port(c)
		addr := fmt.Sprintf("%s://%s:%d", network, host, port)
		state := "open"

		logger := ioutil.Discard
		if c.Bool("verbose") {
			logger = c.App.Writer
		}

		portCheck := portCheckProvider(port).OnHost(host).ForNetwork(network).WithLogger(logger)
		checkFunc := portCheck.IsOpen

		if c.Bool("closed") {
			checkFunc = portCheck.IsClosed
			state = "closed"
		}

		fmt.Fprintf(c.App.Writer, "Waiting for %s to be %s...\n", addr, state)

		if err := waitForConditionWithTimeout(checkFunc, interval, timeout); err != nil {
			fmt.Fprintf(c.App.Writer, "Error waiting for %s port: %s\n", state, err)
			return err
		}

		fmt.Fprintf(c.App.Writer, "Success: port is %s\n", state)
		return nil
	},
}
