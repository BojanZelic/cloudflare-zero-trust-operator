package cfcompare_test

import (
	"testing"

	. "github.com/bojanzelic/cloudflare-zero-trust-operator/tests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPIs(t *testing.T) {
	_ = TestFlags // import mandatory, so that `go test` do not complain about `flag provided but not defined`
	RegisterFailHandler(Fail)

	RunSpecs(t, "CfCollections Suite")
}
