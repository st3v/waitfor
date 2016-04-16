package check_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var fakeBin string

func TestCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "check")
}

var _ = BeforeSuite(func() {
	var err error
	fakeBin, err = gexec.Build("github.com/st3v/waitfor/check/fake/command")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
