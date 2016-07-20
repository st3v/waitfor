package main

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/st3v/waitfor"
	"github.com/st3v/waitfor/check"
	"github.com/st3v/waitfor/cmd/waitfor/fake"
)

var _ = Describe("port command", func() {
	var (
		app = app()

		portcheck   *fake.PortCheck
		expectedErr error
		args        []string

		expectedPort   int
		actualPort     int
		actualInterval time.Duration
		actualTimeout  time.Duration
		actualOutput   *gbytes.Buffer
	)

	BeforeEach(func() {
		portcheck = new(fake.PortCheck)
		portcheck.OnHostReturns(portcheck)
		portcheck.ForNetworkReturns(portcheck)
		portcheck.WithLoggerReturns(portcheck)
		portCheckProvider = func(port int) check.PortCheck {
			actualPort = port
			return portcheck
		}

		args = []string{}
		expectedPort = 12345

		actualPort = 0
		actualInterval = 0
		actualTimeout = 0
		actualOutput = gbytes.NewBuffer()

		waitForConditionWithTimeout = func(check waitfor.Check, interval, timeout time.Duration) error {
			check()
			actualInterval = interval
			actualTimeout = timeout
			return expectedErr
		}
	})

	JustBeforeEach(func() {
		expectedErr = nil
		app.Writer = io.MultiWriter(GinkgoWriter, actualOutput)
		args = append([]string{"waitfor", "port", strconv.Itoa(expectedPort)}, args...)
		app.Run(args)
	})

	Context("when the check succeeds", func() {
		It("logs a success message", func() {
			Expect(actualOutput).To(gbytes.Say("Success"))
		})
	})

	Context("when the check fails", func() {
		JustBeforeEach(func() {
			expectedErr = errors.New("some-error")
		})

		It("returns an error", func() {
			err := app.Run([]string{"watchfor", "port", "123"})
			Expect(err).To(HaveOccurred())
		})

		It("logs the error", func() {
			app.Run([]string{"watchfor", "port", "123"})
			Expect(actualOutput).To(gbytes.Say(expectedErr.Error()))
		})
	})

	Describe("port argument", func() {
		It("checks the specified port", func() {
			Expect(actualPort).To(Equal(expectedPort))
		})

		Context("when the port has not been specified", func() {
			var exitCode int

			JustBeforeEach(func() {
				exitCode = 0
				exit = func(rc int) {
					exitCode = rc
					panic(rc)
				}

				Expect(func() {
					app.Run([]string{"watchfor", "port"})
				}).To(Panic())
			})

			AfterEach(func() {
				exit = os.Exit
			})

			It("exits with non-zero exit code", func() {
				Expect(exitCode).ToNot(Equal(0))
			})

			It("provides a corresponding error", func() {
				Expect(actualOutput).To(gbytes.Say("must specify port"))
			})
		})

		Context("when the port is invalid", func() {
			var exitCode int

			JustBeforeEach(func() {
				exitCode = 0
				exit = func(rc int) {
					exitCode = rc
					panic(rc)
				}

				Expect(func() {
					app.Run([]string{"watchfor", "port", "invalid"})
				}).To(Panic())
			})

			AfterEach(func() {
				exit = os.Exit
			})

			It("exits with non-zero exit code", func() {
				Expect(exitCode).ToNot(Equal(0))
			})

			It("provides a corresponding error", func() {
				Expect(actualOutput).To(gbytes.Say("invalid port"))
			})
		})
	})

	Describe("--closed flag", func() {
		Context("when it has been specified", func() {
			BeforeEach(func() {
				args = []string{"--closed"}
			})

			It("uses the IsClosed check", func() {
				Expect(portcheck.IsClosedCallCount()).To(Equal(1))
			})

			It("logs the correct state", func() {
				Expect(actualOutput).To(gbytes.Say("to be closed"))
				Expect(actualOutput).To(gbytes.Say("Success: port is closed"))
			})
		})

		Context("when it has not been specified", func() {
			It("uses the IsOpen check", func() {
				Expect(portcheck.IsOpenCallCount()).To(Equal(1))
			})

			It("logs the correct state", func() {
				Expect(actualOutput).To(gbytes.Say("to be open"))
				Expect(actualOutput).To(gbytes.Say("Success: port is open"))
			})
		})
	})

	Describe("--host flag", func() {
		Context("when it has been set", func() {
			var expectedHost = "1.2.3.4"

			BeforeEach(func() {
				args = []string{"--host", expectedHost}
			})

			It("checks the specified host", func() {
				Expect(portcheck.OnHostArgsForCall(0)).To(Equal(expectedHost))
			})
		})

		Context("when it has not been set", func() {
			It("checks the default host", func() {
				Expect(portcheck.OnHostArgsForCall(0)).To(Equal("127.0.0.1"))
			})
		})
	})

	Describe("--network flag", func() {
		Context("when it has been set", func() {
			var expectedNetwork = "udp"

			BeforeEach(func() {
				args = []string{"--network", expectedNetwork}
			})

			It("checks the specified network", func() {
				Expect(portcheck.ForNetworkArgsForCall(0)).To(Equal(expectedNetwork))
			})
		})

		Context("when it has not been set", func() {
			It("checks the default network", func() {
				Expect(portcheck.ForNetworkArgsForCall(0)).To(Equal("tcp"))
			})
		})
	})

	Describe("--verbose flag", func() {
		Context("when it has been set", func() {
			BeforeEach(func() {
				args = []string{"--verbose"}
			})

			It("is uses the correct writer for logging", func() {
				Expect(portcheck.WithLoggerArgsForCall(0)).To(Equal(app.Writer))
			})
		})

		Context("it uses io.Discard for logging", func() {
			It("", func() {
				Expect(portcheck.WithLoggerArgsForCall(0)).To(Equal(ioutil.Discard))
			})
		})
	})

	Describe("--interval flag", func() {
		Context("when it has been set", func() {
			var expectedInterval = 123 * time.Second

			BeforeEach(func() {
				args = []string{"--interval", expectedInterval.String()}
			})

			It("is being used", func() {
				Expect(actualInterval).To(Equal(expectedInterval))
			})
		})

		Context("when it has not been set", func() {
			It("the default interval is being used", func() {
				Expect(actualInterval).To(Equal(time.Second))
			})
		})
	})

	Describe("--timeout flag", func() {
		Context("when it has been set", func() {
			var expectedTimeout = 123 * time.Second

			BeforeEach(func() {
				args = []string{"--timeout", expectedTimeout.String()}
			})

			It("is being used", func() {
				Expect(actualTimeout).To(Equal(expectedTimeout))
			})
		})

		Context("when it has not been set", func() {
			It("the default timeout is being used", func() {
				Expect(actualTimeout).To(Equal(300 * time.Second))
			})
		})
	})
})
