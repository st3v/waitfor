package check_test

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
	"github.com/st3v/waitfor/check"
)

var _ = Describe("curlcheck", func() {
	var (
		server    *ghttp.Server
		curlcheck check.CurlCheck
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		methods := []string{"HEAD", "GET", "POST", "PUT", "DELETE"}
		for _, method := range methods {
			server.RouteToHandler(method, "/success", ghttp.RespondWith(200, "success"))
			server.RouteToHandler(method, "/missing", ghttp.RespondWith(404, "missing"))
		}
	})

	DescribeTable(".MatchResponseCode",
		func(path, method string, statusCode int, expected bool) {
			curlcheck = check.Curl(fmt.Sprintf("%s/%s", server.URL(), path)).WithLogger(GinkgoWriter)

			if method != "" {
				curlcheck.WithMethod(method)
			} else {
				method = check.DefaultCurlMethod
			}

			match := curlcheck.MatchResponseCode(statusCode)
			Expect(match).To(Equal(expected))

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(server.ReceivedRequests()[0].Method).To(Equal(method))
		},
		Entry("matches HEAD /success and HTTP 200", "success", "", 200, true),
		Entry("does not match HEAD /success and HTTP 404", "success", "", 404, false),
		Entry("matches HEAD /missing and HTTP 404", "missing", "", 404, true),
		Entry("does not match HEAD /missing and HTTP 200", "missing", "", 200, false),
		Entry("matches HEAD /success and HTTP 200", "success", "HEAD", 200, true),
		Entry("does not match HEAD /success and HTTP 404", "success", "HEAD", 404, false),
		Entry("matches HEAD /missing and HTTP 404", "missing", "HEAD", 404, true),
		Entry("does not match HEAD /missing and HTTP 200", "missing", "HEAD", 200, false),
		Entry("matches GET /success and HTTP 200", "success", "", 200, true),
		Entry("does not match GET /success and HTTP 404", "success", "", 404, false),
		Entry("matches GET /missing and HTTP 404", "missing", "", 404, true),
		Entry("does not match GET /missing and HTTP 200", "missing", "", 200, false),
		Entry("matches POST /success and HTTP 200", "success", "", 200, true),
		Entry("does not match POST /success and HTTP 404", "success", "", 404, false),
		Entry("matches POST /missing and HTTP 404", "missing", "", 404, true),
		Entry("does not match POST /missing and HTTP 200", "missing", "", 200, false),
		Entry("matches PUT /success and HTTP 200", "success", "", 200, true),
		Entry("does not match PUT /success and HTTP 404", "success", "", 404, false),
		Entry("matches PUT /missing and HTTP 404", "missing", "", 404, true),
		Entry("does not match PUT /missing and HTTP 200", "missing", "", 200, false),
		Entry("matches DELETE /success and HTTP 200", "success", "", 200, true),
		Entry("does not match DELETE /success and HTTP 404", "success", "", 404, false),
		Entry("matches DELETE /missing and HTTP 404", "missing", "", 404, true),
		Entry("does not match DELETE /missing and HTTP 200", "missing", "", 200, false),
	)

	DescribeTable(".MatchBody",
		func(path, method, regex string, expected bool) {
			curlcheck = check.Curl(fmt.Sprintf("%s/%s", server.URL(), path)).WithLogger(GinkgoWriter)

			if method != "" {
				curlcheck.WithMethod(method)
			} else {
				method = check.DefaultCurlMethod
			}

			match := curlcheck.MatchBody(regexp.MustCompile(regex))
			Expect(match).To(Equal(expected))

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(server.ReceivedRequests()[0].Method).To(Equal(method))
		},
		Entry("does not match HEAD /success and regex '.*success.*'", "success", "", ".*success.*", false),
		Entry("does not match HEAD /success and regex '.*error.*'", "success", "", ".*error.*", false),
		Entry("does not match HEAD /missing and regex '.*success.*'", "missing", "", ".*success.*", false),
		Entry("does not match HEAD /missing and regex '.*missing.*'", "missing", "", ".*missing.*", false),
		Entry("does not match HEAD /success and regex '.*success.*'", "success", "HEAD", ".*success.*", false),
		Entry("does not match HEAD /success and regex '.*error.*'", "success", "HEAD", ".*success.*", false),
		Entry("does not match HEAD /missing and regex '.*success.*'", "missing", "HEAD", ".*success.*", false),
		Entry("does not match HEAD /missing and regex '.*missing.*'", "missing", "HEAD", ".*missing.*", false),
		Entry("matches GET /success and regex '.*success.*'", "success", "GET", ".*success.*", true),
		Entry("does not match GET /success and regex '.*error.*'", "success", "GET", ".*error.*", false),
		Entry("does not match GET /missing and regex '.*success.*'", "missing", "GET", ".*success.*", false),
		Entry("matches GET /missing and regex '.*missing.*'", "missing", "GET", ".*missing.*", true),
		Entry("matches POST /success and regex '.*success.*'", "success", "POST", ".*success.*", true),
		Entry("does not match POST /success and regex '.*error.*'", "success", "POST", ".*error.*", false),
		Entry("does not match POST /missing and regex '.*success.*'", "missing", "POST", ".*success.*", false),
		Entry("matches POST /missing and regex '.*missing.*'", "missing", "POST", ".*missing.*", true),
		Entry("matches PUT /success and regex '.*success.*'", "success", "PUT", ".*success.*", true),
		Entry("does not match PUT /success and regex '.*error.*'", "success", "PUT", ".*error.*", false),
		Entry("does not match PUT /missing and regex '.*success.*'", "missing", "PUT", ".*success.*", false),
		Entry("matches PUT /missing and regex '.*missing.*'", "missing", "PUT", ".*missing.*", true),
		Entry("matches DELETE /success and regex '.*success.*'", "success", "DELETE", ".*success.*", true),
		Entry("does not match DELETE /success and regex '.*error.*'", "success", "DELETE", ".*error.*", false),
		Entry("does not match DELETE /missing and regex '.*success.*'", "missing", "DELETE", ".*success.*", false),
		Entry("matches DELETE /missing and regex '.*missing.*'", "missing", "DELETE", ".*missing.*", true),
	)

	Context("when the server requires authentication", func() {
		var (
			username = "user"
			password = "pass"
		)

		BeforeEach(func() {
			server.RouteToHandler("GET", "/auth", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth(username, password),
				ghttp.RespondWith(200, "authenticated"),
			))
			curlcheck = check.Curl(fmt.Sprintf("%s/auth", server.URL())).WithMethod("GET").WithLogger(GinkgoWriter)
		})

		Context("and the correct credentials have been specified", func() {
			BeforeEach(func() {
				curlcheck.WithAuth(username, password)
			})

			Describe(".MatchResponseCode", func() {
				It("returns true", func() {
					match := curlcheck.MatchResponseCode(200)
					Expect(match).To(BeTrue())
				})
			})

			Describe(".MatchBody", func() {
				It("returns true", func() {
					r := regexp.MustCompile("authenticated")
					match := curlcheck.MatchBody(r)
					Expect(match).To(BeTrue())
				})
			})
		})
	})

	Context("when the server requires a body", func() {
		var data = "expected body"

		BeforeEach(func() {
			server.RouteToHandler("GET", "/header", ghttp.CombineHandlers(
				ghttp.VerifyBody([]byte(data)),
				ghttp.RespondWith(200, "success"),
			))
			curlcheck = check.Curl(fmt.Sprintf("%s/header", server.URL())).WithMethod("GET").WithLogger(GinkgoWriter)
		})

		Context("and the correct headers have been specified", func() {
			BeforeEach(func() {
				curlcheck.WithData(strings.NewReader(data))
			})

			Describe(".MatchResponseCode", func() {
				It("returns true", func() {
					match := curlcheck.MatchResponseCode(200)
					Expect(match).To(BeTrue())
				})
			})

			Describe(".MatchBody", func() {
				It("returns true", func() {
					r := regexp.MustCompile("success")
					match := curlcheck.MatchBody(r)
					Expect(match).To(BeTrue())
				})
			})
		})
	})

	Context("when the server requires certain headers", func() {
		var headers = map[string]string{
			"foo":          "bar",
			"Content-Type": "application/json",
		}

		BeforeEach(func() {
			server.RouteToHandler("GET", "/header", ghttp.CombineHandlers(
				ghttp.VerifyHeader(http.Header{
					"foo": []string{"bar"},
				}),
				ghttp.VerifyContentType("application/json"),
				ghttp.RespondWith(200, "success"),
			))
			curlcheck = check.Curl(fmt.Sprintf("%s/header", server.URL())).WithMethod("GET").WithLogger(GinkgoWriter)
		})

		Context("and the correct headers have been specified", func() {
			BeforeEach(func() {
				for k, v := range headers {
					curlcheck.WithHeader(k, v)
				}
			})

			Describe(".MatchResponseCode", func() {
				It("returns true", func() {
					match := curlcheck.MatchResponseCode(200)
					Expect(match).To(BeTrue())
				})
			})

			Describe(".MatchBody", func() {
				It("returns true", func() {
					r := regexp.MustCompile("success")
					match := curlcheck.MatchBody(r)
					Expect(match).To(BeTrue())
				})
			})
		})
	})

	Context("when a logger is being passed", func() {
		var output *gbytes.Buffer

		BeforeEach(func() {
			output = gbytes.NewBuffer()
			logger := io.MultiWriter(GinkgoWriter, output)
			curlcheck = check.Curl(fmt.Sprintf("%s/success", server.URL())).WithMethod("GET").WithLogger(logger)
		})

		Describe(".MatchResponseCode", func() {
			It("provides logging", func() {
				curlcheck.MatchResponseCode(200)
				Expect(output).To(gbytes.Say(fmt.Sprintf("curl GET %s/success ...", server.URL())))
				Expect(output).To(gbytes.Say("got HTTP status code 200"))
			})
		})

		Describe(".MatchBody", func() {
			It("provides logging", func() {
				r := regexp.MustCompile("success")
				curlcheck.MatchBody(r)
				Expect(output).To(gbytes.Say(fmt.Sprintf("curl GET %s/success ...", server.URL())))
				Expect(output).To(gbytes.Say("got HTTP status code 200 and body:\nsuccess"))
			})
		})
	})

	Context("when a connection error occurs", func() {
		BeforeEach(func() {
			port, err := freeTcpPort()
			Expect(err).ToNot(HaveOccurred())

			addr := fmt.Sprintf("http://localhost:%d", port)
			curlcheck = check.Curl(fmt.Sprintf("%s", addr)).WithLogger(GinkgoWriter)
		})

		Describe(".MatchResponseCode", func() {
			It("returns false", func() {
				match := curlcheck.MatchResponseCode(200)
				Expect(match).To(BeFalse())
			})
		})

		Describe(".MatchBody", func() {
			It("returns false", func() {
				r := regexp.MustCompile("success")
				match := curlcheck.MatchBody(r)
				Expect(match).To(BeFalse())
			})
		})
	})

	Context("when an invalid url has been specified", func() {
		var output *gbytes.Buffer

		BeforeEach(func() {
			output = gbytes.NewBuffer()
			logger := io.MultiWriter(GinkgoWriter, output)
			curlcheck = check.Curl("%invalid-url@").WithLogger(logger)
		})

		Describe(".MatchResponseCode", func() {
			It("returns false", func() {
				match := curlcheck.MatchResponseCode(200)
				Expect(match).To(BeFalse())
			})

			It("logs the error", func() {
				curlcheck.MatchResponseCode(200)
				Expect(output).To(gbytes.Say("invalid URL escape"))
			})
		})

		Describe(".MatchBody", func() {
			It("returns false", func() {
				r := regexp.MustCompile("success")
				match := curlcheck.MatchBody(r)
				Expect(match).To(BeFalse())
			})

			It("logs the error", func() {
				curlcheck.MatchResponseCode(200)
				Expect(output).To(gbytes.Say("invalid URL escape"))
			})
		})
	})
})
