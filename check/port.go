package check

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
)

var (
	DefaultHost    = "127.0.0.1"
	DefaultNetwork = "tcp"
	DefaultLogger  = ioutil.Discard
)

type PortCheck interface {
	IsOpen() bool
	IsClosed() bool
	OnHost(string) PortCheck
	ForNetwork(string) PortCheck
	WithLogger(io.Writer) PortCheck
}

type portcheck struct {
	port    int
	host    string
	network string
	logger  io.Writer
}

func Port(p int) PortCheck {
	return &portcheck{
		port:    p,
		host:    DefaultHost,
		network: DefaultNetwork,
		logger:  DefaultLogger,
	}
}

func (p *portcheck) OnHost(host string) PortCheck {
	p.host = host
	return p
}

func (p *portcheck) ForNetwork(network string) PortCheck {
	p.network = network
	return p
}

func (p *portcheck) WithLogger(w io.Writer) PortCheck {
	p.logger = w
	return p
}

func (p *portcheck) IsOpen() bool {
	fmt.Fprintf(p.logger, "Dialing %s://%s\n", p.network, p.addr())

	conn, err := net.Dial(p.network, p.addr())
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

func (p *portcheck) IsClosed() bool {
	return !p.IsOpen()
}

func (p *portcheck) addr() string {
	return fmt.Sprintf("%s:%d", p.host, p.port)
}
