package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLinuxCircusSoldier(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linux-Circus-Soldier Suite")
}
