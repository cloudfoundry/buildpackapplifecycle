package smelter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSmelter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Smelter Suite")
}
