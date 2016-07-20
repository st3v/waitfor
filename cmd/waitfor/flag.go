package main

import (
	"time"

	"github.com/codegangsta/cli"
)

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

var failFlag = cli.BoolFlag{
	Name:  "fail, f",
	Usage: "wait for condition to fail",
}

var dataFlag = cli.StringSliceFlag{
	Name:  "data, d",
	Value: &cli.StringSlice{},
	Usage: "HTTP POST data",
}

var headerFlag = cli.StringSliceFlag{
	Name:  "header, H",
	Value: &cli.StringSlice{},
	Usage: "custom header",
}

var userFlag = cli.StringFlag{
	Name:  "user, u",
	Value: "",
	Usage: "username and password separated by colon, e.g. 'username:password'",
}

var methodFlag = cli.StringFlag{
	Name:  "request, X",
	Value: "GET",
	Usage: "method for curl request",
}

var httpStatusFlag = cli.StringFlag{
	Name:  "status, s",
	Value: "200",
	Usage: "match HTTP status code for curl request",
}

var exitCodeFlag = cli.StringFlag{
	Name:  "status, rc",
	Value: "0",
	Usage: "match exit code",
}

var matchFlag = cli.StringFlag{
	Name:  "match, m",
	Value: "",
	Usage: "match regex",
}
