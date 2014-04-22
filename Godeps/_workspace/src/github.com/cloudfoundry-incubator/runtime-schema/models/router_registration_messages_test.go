package models_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("RouterRegistrationMessages", func() {
	var ccJSON = `{
        "host": "127.0.0.1",
        "port": 4567,
        "tags": {"component":"CloudController"}
      }`

	var registrationMessage RouterRegistrationMessage

	BeforeEach(func() {
		err := json.Unmarshal([]byte(ccJSON), &registrationMessage)
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("should be mapped to the CC's registration JSON", func() {
		Ω(registrationMessage).Should(Equal(RouterRegistrationMessage{
			Host: "127.0.0.1",
			Port: 4567,
			Tags: map[string]string{"component": "CloudController"},
		}))
	})

	Describe("fetching the component", func() {
		It("should return the component when present", func() {
			Ω(registrationMessage.Component()).Should(Equal("CloudController"))
		})
	})
})
