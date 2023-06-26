package test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFormulator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Formulator Suite")
}
