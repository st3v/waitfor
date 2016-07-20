package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/codegangsta/cli"

	"github.com/st3v/waitfor/check"
)

var curlCheckProvider = check.Curl

var url = func(c *cli.Context) string {
	if !c.Args().Present() {
		cli.ShowCommandHelp(c, "curl")
		fmt.Fprintln(c.App.Writer, "must specify url")
		exit(1)
	}
	return c.Args().First()
}

func splitByColon(str string) (string, string) {
	parts := strings.Split(str, ":")

	one := parts[0]
	two := ""

	if len(parts) > 1 {
		two = parts[1]
	}

	return one, two
}

var curlCommand = cli.Command{
	Name:  "curl",
	Usage: "wait for curl to succeed (or fail)",

	HideHelp: true,

	Flags: []cli.Flag{
		httpStatusFlag,
		matchFlag,
		methodFlag,
		userFlag,
		dataFlag,
		headerFlag,
		failFlag,
		timeoutFlag,
		intervalFlag,
		verboseFlag,
	},

	Action: func(c *cli.Context) {
		statusCode := c.Int("status")
		regex := c.String("match")
		method := c.String("method")
		data := c.StringSlice("data")
		headers := append(c.StringSlice("header"), fmt.Sprintf("user-agent:waitfor/%s", c.App.Version))
		auth := c.String("user")
		negate := c.Bool("fail")
		timeout := c.Duration("timeout")
		interval := c.Duration("interval")
		verbose := c.Bool("verbose")

		logger := ioutil.Discard
		if verbose {
			logger = c.App.Writer
		}

		curlCheck := curlCheckProvider(url(c)).WithMethod(method).WithLogger(logger)

		if auth != "" {
			curlCheck.WithAuth(splitByColon(auth))
		}

		for _, h := range headers {
			curlCheck.WithHeader(splitByColon(h))
		}

		if len(data) > 0 {
			curlCheck.WithData(strings.NewReader(strings.Join(data, "&")))
		}

		condition := func() bool {
			return curlCheck.MatchResponseCode(statusCode)
		}

		if regex != "" {
			r := regexp.MustCompile(regex)
			condition = func() bool {
				return curlCheck.MatchBody(r)
			}
		}

		state := "succeed"
		if negate {
			state = "fail"
			aux := condition
			condition = func() bool {
				return !aux()
			}
		}

		fmt.Fprintf(c.App.Writer, "Waiting for curl to %s...\n", state)
		if err := waitForConditionWithTimeout(condition, interval, timeout); err != nil {
			fmt.Fprintf(c.App.Writer, "Error waiting for curl to %s: %s\n", state, err)
			exit(1)
		}

		fmt.Fprintf(c.App.Writer, "Success: curl did %s\n", state)
	},
}
