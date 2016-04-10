package check

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

var (
	DefaultCurlMethod = http.MethodHead
)

type CurlCheck interface {
	MatchResponseCode(statusCode int) bool
	MatchBody(regex *regexp.Regexp) bool

	WithMethod(string) CurlCheck
	WithAuth(string, string) CurlCheck
	WithHeader(string, string) CurlCheck
	WithData(io.Reader) CurlCheck
	WithLogger(io.Writer) CurlCheck
}

type curlcheck struct {
	url      string
	method   string
	username string
	password string
	headers  map[string]string
	data     io.Reader
	logger   io.Writer
}

func Curl(url string) CurlCheck {
	return &curlcheck{
		url:     url,
		method:  DefaultCurlMethod,
		headers: map[string]string{},
		logger:  DefaultLogger,
	}
}

func (c *curlcheck) WithMethod(method string) CurlCheck {
	c.method = method
	return c
}

func (c *curlcheck) WithAuth(username, password string) CurlCheck {
	c.username = username
	c.password = password
	return c
}

func (c *curlcheck) WithHeader(key, value string) CurlCheck {
	c.headers[key] = value
	return c
}

func (c *curlcheck) WithData(data io.Reader) CurlCheck {
	c.data = data
	return c
}

func (c *curlcheck) WithLogger(w io.Writer) CurlCheck {
	c.logger = w
	return c
}

type matcher func(*http.Response, []byte) bool

func (c *curlcheck) MatchBody(regex *regexp.Regexp) bool {
	matcher := func(resp *http.Response, body []byte) bool {
		return regex.Match(body)
	}

	return c.matchResponse(matcher)
}

func (c *curlcheck) MatchResponseCode(statusCode int) bool {
	matcher := func(resp *http.Response, body []byte) bool {
		return resp.StatusCode == statusCode
	}

	return c.matchResponse(matcher)
}

func (c *curlcheck) matchResponse(m matcher) bool {
	resp, err := c.response()
	if err != nil {
		fmt.Fprintln(c.logger, err.Error())
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(c.logger, err.Error())
		return false
	}

	fmt.Fprintf(c.logger, "got HTTP status code %d and body:\n%s\n", resp.StatusCode, string(body))

	return m(resp, body)
}

func (c *curlcheck) response() (*http.Response, error) {
	req, err := c.request()
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(c.logger, "curl %s %s ...\n", c.method, c.url)
	return http.DefaultClient.Do(req)
}

func (c *curlcheck) request() (*http.Request, error) {
	req, err := http.NewRequest(c.method, c.url, c.data)
	if err != nil {
		return nil, err
	}

	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
