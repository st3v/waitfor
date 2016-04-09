package check_test

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"

	"github.com/st3v/waitfor/check"
)

var _ = Describe("Port", func() {
	var (
		logger    *gbytes.Buffer
		host      string
		network   string
		port      int
		result    bool
		portcheck check.PortCheck
	)

	BeforeEach(func() {
		logger = gbytes.NewBuffer()
		host = "localhost"
		network = "tcp"
	})

	Describe(".IsOpen", func() {
		Context("when the port is open", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = ghttp.NewServer()
				host, port = hostPort(server)
			})

			JustBeforeEach(func() {
				portcheck = check.Port(port).OnHost(host)
				result = portcheck.IsOpen()
			})

			AfterEach(func() {
				server.Close()
			})

			It("returns true", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("when the port is closed", func() {
			BeforeEach(func() {
				var err error
				port, err = freeTcpPort()
				Expect(err).ToNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				portcheck = check.Port(port).OnHost(host)
				result = portcheck.IsOpen()
			})

			It("returns false", func() {
				Expect(result).To(BeFalse())
			})
		})

		Context("when host has not been specified", func() {
			BeforeEach(func() {
				portcheck = check.Port(port).ForNetwork(network).WithLogger(logger)
				result = portcheck.IsOpen()
			})

			It("uses the default host", func() {
				expected := fmt.Sprintf("Dialing %s://%s:%d", network, check.DefaultHost, port)
				Expect(logger).To(gbytes.Say(expected))
			})
		})

		Context("when network has not been specified", func() {
			BeforeEach(func() {
				portcheck = check.Port(port).OnHost(host).WithLogger(logger)
				result = portcheck.IsOpen()
			})

			It("uses the default network", func() {
				expected := fmt.Sprintf("Dialing %s://%s:%d", check.DefaultNetwork, host, port)
				Expect(logger).To(gbytes.Say(expected))
			})
		})
	})

	Describe(".IsClosed", func() {
		Context("when the port is open", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = ghttp.NewServer()
				host, port = hostPort(server)
			})

			JustBeforeEach(func() {
				portcheck = check.Port(port).OnHost(host)
				result = portcheck.IsClosed()
			})

			AfterEach(func() {
				server.Close()
			})

			It("returns false", func() {
				Expect(result).To(BeFalse())
			})
		})

		Context("when the port is closed", func() {
			BeforeEach(func() {
				var err error
				port, err = freeTcpPort()
				Expect(err).ToNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				portcheck = check.Port(port).OnHost(host)
				result = portcheck.IsClosed()
			})

			It("returns true", func() {
				Expect(result).To(BeTrue())
			})
		})
	})

	Describe("logging", func() {
		BeforeEach(func() {
			portcheck = check.Port(port).OnHost(host).ForNetwork(network).WithLogger(logger)
			result = portcheck.IsOpen()
		})

		It("provides logging", func() {
			expected := fmt.Sprintf("Dialing %s://%s:%d", network, host, port)
			Expect(logger).To(gbytes.Say(expected))
		})
	})
})

func freeTcpPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

func hostPort(server *ghttp.Server) (host string, port int) {
	url, err := url.Parse(server.URL())
	Expect(err).ToNot(HaveOccurred())

	var portStr string
	host, portStr, err = net.SplitHostPort(url.Host)
	Expect(err).ToNot(HaveOccurred())

	port, err = strconv.Atoi(portStr)
	Expect(err).ToNot(HaveOccurred())
	return
}
