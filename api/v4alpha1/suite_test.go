package v4alpha1_test

import (
	"testing"

	. "github.com/bojanzelic/cloudflare-zero-trust-operator/tests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	// +kubebuilder:scaffold:imports
)

func TestK8SAPI(t *testing.T) {
	_ = TestFlags // import mandatory, so that `go test` do not complain about `flag provided but not defined`
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubebuilder API Suite")
}
