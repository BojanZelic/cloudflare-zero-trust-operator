package v4alpha1_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	// +kubebuilder:scaffold:imports
)

func TestK8SAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubebuilder API Suite")
}
