package ctrlhelper_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	// +kubebuilder:scaffold:imports
)

func TestControllerHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubebuilder Controllers Helpers")
}
