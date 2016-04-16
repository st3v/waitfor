package check

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

type CommandCheck interface {
	Succeeds() bool
	Fails() bool
	MatchesExitCode(int) bool
	MatchesOutput(*regexp.Regexp) bool

	WithEnv([]string) CommandCheck
	WithLogger(io.Writer) CommandCheck
	WithStdin(io.Reader) CommandCheck
}

type cmdcheck struct {
	cmd    string
	args   []string
	env    []string
	stdin  io.Reader
	logger io.Writer
}

func Command(cmd string, args ...string) CommandCheck {
	return &cmdcheck{
		cmd:    cmd,
		args:   args,
		logger: DefaultLogger,
	}
}

func (c *cmdcheck) WithEnv(env []string) CommandCheck {
	c.env = env
	return c
}

func (c *cmdcheck) WithLogger(w io.Writer) CommandCheck {
	c.logger = w
	return c
}

func (c *cmdcheck) WithStdin(r io.Reader) CommandCheck {
	c.stdin = r
	return c
}

func (c *cmdcheck) Succeeds() bool {
	_, err := c.exec()
	return err == nil
}

func (c *cmdcheck) Fails() bool {
	return !c.Succeeds()
}

func (c *cmdcheck) MatchesOutput(regex *regexp.Regexp) bool {
	out, _ := c.exec()
	return regex.Match(out)
}

func (c *cmdcheck) MatchesExitCode(exitCode int) bool {
	_, err := c.exec()

	rc := 0

	switch err := err.(type) {
	case *exec.ExitError:
		waitStatus := err.Sys().(syscall.WaitStatus)
		rc = waitStatus.ExitStatus()
	case error:
		rc = 127
	}

	return rc == (exitCode%256+256)%256
}

func (c *cmdcheck) exec() ([]byte, error) {
	msg := strings.Join(append([]string{"Running", c.cmd}, c.args...), " ")
	fmt.Fprintln(c.logger, msg)

	cmd := exec.Command(c.cmd, c.args...)

	if len(c.env) > 0 {
		cmd.Env = c.env
	}

	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}

	out, err := cmd.CombinedOutput()

	if len(out) > 0 {
		fmt.Fprint(c.logger, string(out))
	}

	if err != nil {
		fmt.Fprintln(c.logger, err.Error())
	}

	return out, err
}
