package check_test

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/st3v/waitfor/check"
)

var _ = Describe("cmdcheck", func() {
	DescribeTable(".Succeeds",
		func(args []string, expected bool) {
			command := check.Command(fakeBin, args...).WithLogger(GinkgoWriter)
			Expect(command.Succeeds()).To(Equal(expected))
		},
		Entry("returns true when exit code is 0",
			[]string{"--exit", "0"}, true,
		),
		Entry("returns false when exit code 1",
			[]string{"--exit", "1"}, false,
		),
		Entry("returns true when exit code is 0 but stderr has content",
			[]string{"--exit", "0", "--err", "Error"}, true,
		),
	)

	DescribeTable(".Fails",
		func(args []string, expected bool) {
			command := check.Command(fakeBin, args...).WithLogger(GinkgoWriter)
			Expect(command.Fails()).To(Equal(expected))
		},
		Entry("returns true when exit code is 1", []string{"--exit", "1"}, true),
		Entry("returns false when exit code 0", []string{"--exit", "0"}, false),
	)

	DescribeTable(".MatchesExitCode",
		func(args []string, match int, expected bool) {
			command := check.Command(fakeBin, args...).WithLogger(GinkgoWriter)
			Expect(command.MatchesExitCode(match)).To(Equal(expected))
		},
		Entry("matches exit code 0", []string{"--exit", "0"}, 0, true),
		Entry("does not match exit code 1", []string{"--exit", "0"}, 1, false),
		Entry("matches exit code 1", []string{"--exit", "1"}, 1, true),
		Entry("does not match exit code 0", []string{"--exit", "1"}, 0, false),
		Entry("matches exit code -1", []string{"--exit", "-1"}, -1, true),
		Entry("matches exit code -256", []string{"--exit", "-256"}, -256, true),
		Entry("matches exit code -1001", []string{"--exit", "-1001"}, -1001, true),
	)

	DescribeTable(".MatchesOutput",
		func(args []string, match string, expected bool) {
			regex := regexp.MustCompile(match)
			command := check.Command(fakeBin, args...).WithLogger(GinkgoWriter)
			Expect(command.MatchesOutput(regex)).To(Equal(expected))
		},
		Entry("returns true when there is a match on stdout",
			[]string{"--out", "Hello World"}, ".*ell.*", true,
		),
		Entry("returns true when there is a match on stderr",
			[]string{"--err", "This is an error"}, ".*err.*", true,
		),
		Entry("returns false when there is no match on stdout and stderr",
			[]string{"--out", "Hello", "--err", "Error"}, ".*\\sell.*", false,
		),
		Entry("returns false when both stdout and stderr are empty",
			[]string{}, ".+", false,
		),
		Entry("returns true when the regex matches empty strings",
			[]string{}, ".*", true,
		),
		Entry("returns true when there is a match and exit code non-zero",
			[]string{"--out", "Hello World", "--exit", "1"}, ".*ell.*", true,
		),
	)

	Context("when the command cannot be executed", func() {
		DescribeTable("cmdcheck",
			func(fn func(check.CommandCheck) bool, expected bool) {
				command := check.Command("not-a-valid-command").WithLogger(GinkgoWriter)
				Expect(fn(command)).To(Equal(expected))
			},
			Entry(".Succeeds returns false",
				func(cmd check.CommandCheck) bool { return cmd.Succeeds() }, false,
			),
			Entry(".Fails returns true",
				func(cmd check.CommandCheck) bool { return cmd.Fails() }, true,
			),
			Entry(".MatchesExitCode returns false for exit code 0",
				func(cmd check.CommandCheck) bool { return cmd.MatchesExitCode(0) }, false,
			),
			Entry(".MatchesExitCode returns true for exit code 127",
				func(cmd check.CommandCheck) bool { return cmd.MatchesExitCode(127) }, true,
			),
			Entry(".MatchesOutput returns false if regex does not match empty strings",
				func(cmd check.CommandCheck) bool { return cmd.MatchesOutput(regexp.MustCompile(".+")) }, false,
			),
			Entry(".MatchesOutput returns true if the regex matches empty strings",
				func(cmd check.CommandCheck) bool { return cmd.MatchesOutput(regexp.MustCompile(".*")) }, true,
			),
		)
	})

	Context("when env is not being set", func() {
		var (
			output      *gbytes.Buffer
			command     check.CommandCheck
			envVarName  = "FOO"
			envVarValue = "BAR"
		)

		BeforeEach(func() {
			os.Setenv(envVarName, envVarValue)
			output = gbytes.NewBuffer()
			logger := io.MultiWriter(GinkgoWriter, output)
			command = check.Command(fakeBin, "--env").WithLogger(logger)
		})

		DescribeTable("cmdcheck",
			func(fn func(check.CommandCheck) bool, expected bool) {
				Expect(fn(command)).To(Equal(expected))
				Expect(output).To(gbytes.Say(fmt.Sprintf("%s=%s", envVarName, envVarValue)))
			},
			Entry(".Succeeds uses env from parent process",
				func(cmd check.CommandCheck) bool { return cmd.Succeeds() }, true,
			),
			Entry(".Fails uses env from parent process",
				func(cmd check.CommandCheck) bool { return cmd.Fails() }, false,
			),
			Entry(".MatchesExitCode uses env from parent process",
				func(cmd check.CommandCheck) bool { return cmd.MatchesExitCode(0) }, true,
			),
			Entry(".MatchesOutput uses env from parent process",
				func(cmd check.CommandCheck) bool { return cmd.MatchesOutput(regexp.MustCompile(envVarName)) }, true,
			),
		)
	})

	Context("when env is being set", func() {
		var (
			output  *gbytes.Buffer
			command check.CommandCheck
			env     = []string{"ONE=1", "TWO=2"}
		)

		BeforeEach(func() {
			output = gbytes.NewBuffer()
			logger := io.MultiWriter(GinkgoWriter, output)
			command = check.Command(fakeBin, "--env").WithEnv(env).WithLogger(logger)
		})

		DescribeTable("cmdcheck",
			func(fn func(check.CommandCheck) bool, expected bool) {
				Expect(fn(command)).To(Equal(expected))
				for _, v := range env {
					Expect(output).To(gbytes.Say(v))
				}
			},
			Entry(".Succeeds forwards env",
				func(cmd check.CommandCheck) bool { return cmd.Succeeds() }, true,
			),
			Entry(".Fails forwards env",
				func(cmd check.CommandCheck) bool { return cmd.Fails() }, false,
			),
			Entry(".MatchesExitCode forwards env",
				func(cmd check.CommandCheck) bool { return cmd.MatchesExitCode(0) }, true,
			),
			Entry(".MatchesOutput forwards env",
				func(cmd check.CommandCheck) bool { return cmd.MatchesOutput(regexp.MustCompile(env[0])) }, true,
			),
		)
	})

	Context("when stdin is being set", func() {
		var (
			output  *gbytes.Buffer
			command check.CommandCheck
			input   = "Hello World"
		)

		BeforeEach(func() {
			output = gbytes.NewBuffer()
			buf := strings.NewReader(input)
			logger := io.MultiWriter(GinkgoWriter, output)
			command = check.Command(fakeBin, "--echo").WithStdin(buf).WithLogger(logger)
		})

		DescribeTable("cmdcheck",
			func(fn func(check.CommandCheck) bool, expected bool) {
				Expect(fn(command)).To(Equal(expected))
				Expect(output).To(gbytes.Say(input))
			},
			Entry(".Succeeds forwards stdin",
				func(cmd check.CommandCheck) bool { return cmd.Succeeds() }, true,
			),
			Entry(".Fails forwards stdin",
				func(cmd check.CommandCheck) bool { return cmd.Fails() }, false,
			),
			Entry(".MatchesExitCode forwards stdin",
				func(cmd check.CommandCheck) bool { return cmd.MatchesExitCode(0) }, true,
			),
			Entry(".MatchesOutput forwards stdin",
				func(cmd check.CommandCheck) bool { return cmd.MatchesOutput(regexp.MustCompile(input)) }, true,
			),
		)
	})

	Context("when a logger is being passed", func() {
		DescribeTable("cmdcheck",
			func(fn func(check.CommandCheck)) {
				output := gbytes.NewBuffer()
				logger := io.MultiWriter(GinkgoWriter, output)
				args := []string{"--out", "some-output", "--err", "some-error"}
				command := check.Command(fakeBin, args...).WithLogger(logger)

				fn(command)

				Expect(output).To(gbytes.Say(fmt.Sprintf("Running %s %s", fakeBin, strings.Join(args, " "))))
			},
			Entry(".Succeeds provides logging",
				func(cmd check.CommandCheck) { cmd.Succeeds() },
			),
			Entry(".Fails provides logging",
				func(cmd check.CommandCheck) { cmd.Fails() },
			),
			Entry(".MatchesExitCode provides logging",
				func(cmd check.CommandCheck) { cmd.MatchesExitCode(0) },
			),
			Entry(".MatchesOutput provides logging",
				func(cmd check.CommandCheck) { cmd.MatchesOutput(regexp.MustCompile("")) },
			),
		)
	})
})
