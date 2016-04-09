package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

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

var closedFlag = cli.BoolFlag{
	Name:  "closed, c",
	Usage: "wait for port to be closed",
}

var networkFlag = cli.StringFlag{
	Name:  "network, n",
	Value: "tcp",
	Usage: "named network, ['tcp', 'tcp4', 'tcp6', 'udp', 'udp4', 'udp6', 'ip', 'ip4', 'ip6']",
}

var hostFlag = cli.StringFlag{
	Name:  "host, h",
	Value: "127.0.0.1",
	Usage: "resolvable hostname or IP address",
}

var timeoutFlag = cli.DurationFlag{
	Name:  "timeout, t",
	Value: 300 * time.Second,
	Usage: "maximum time to wait for",
}

var intervalFlag = cli.DurationFlag{
	Name:  "interval, i",
	Value: 1 * time.Second,
	Usage: "time in-between checks",
}

var verboseFlag = cli.BoolFlag{
	Name:  "verbose, v",
	Usage: "enable additional logging",
}

var portCommand = cli.Command{
	Name:  "port",
	Usage: "wait for host to listen on port",

	HideHelp: true,

	Flags: []cli.Flag{
		closedFlag,
		hostFlag,
		networkFlag,
		timeoutFlag,
		intervalFlag,
		verboseFlag,
	},

	Action: func(c *cli.Context) {
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
			exit(1)
		}

		fmt.Fprintf(c.App.Writer, "Success: port is %s\n", state)
	},
}
