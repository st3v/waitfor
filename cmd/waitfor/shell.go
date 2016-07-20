package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/codegangsta/cli"

	"github.com/st3v/waitfor"
	"github.com/st3v/waitfor/check"
)

var shellCommand = cli.Command{
	Name:            "sh",
	Aliases:         []string{"shell"},
	Flags:           []cli.Flag{closedFlag},
	Usage:           "wait for arbitrary shell commands to succeed (or fail)",
	SkipFlagParsing: true,
	Action:          shellAction,
}

var shellAction = func(c *cli.Context) {
	timeout := c.GlobalDuration("timeout")
	interval := c.GlobalDuration("interval")
	verbose := c.GlobalBool("verbose")
	fail := c.GlobalBool("fail")
	exitCode := c.GlobalInt("status")
	match := c.GlobalString("match")

	cmd := c.Args().First()
	args := c.Args().Tail()

	// parts := strings.Split(cmd, " ")
	// if len(parts) > 1 {
	// 	cmd = parts[0]
	// 	args = append(parts[1:], args...)
	// }

	command := check.Command(cmd, args...).WithStdin(os.Stdin)

	if verbose {
		command.WithLogger(c.App.Writer)
	}

	checkFunc := command.Succeeds
	state := "succeed"

	if fail {
		checkFunc = command.Fails
		state = "fail"
	}

	if exitCode != 0 {
		checkFunc = func() bool {
			return command.MatchesExitCode(exitCode)
		}
		state = fmt.Sprintf("match exit code %d", exitCode)
	}

	if match != "" {
		r := regexp.MustCompile(match)
		checkFunc = func() bool {
			return command.MatchesOutput(r)
		}

		state = fmt.Sprintf("match regex '%s'", match)
	}

	fmt.Fprintf(c.App.Writer, "Waiting for %s to %s\n", cmd, state)
	if err := waitfor.ConditionWithTimeout(checkFunc, interval, timeout); err != nil {
		fmt.Fprintf(c.App.Writer, "Error waiting for %s: %s\n", cmd, err)
		exit(1)
	}

	fmt.Fprintf(c.App.Writer, "Success: %s did %s\n", cmd, state)
}
