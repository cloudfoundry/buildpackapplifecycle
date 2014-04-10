package models_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("CCRegistrationMessages", func() {
	Describe("CCRegistrationMessage", func() {
		var ccJSON = `{
        "host": "127.0.0.1",
        "port": 4567
      }`

		It("should be mapped to the CC's registration JSON", func() {
			var registrationMessage CCRegistrationMessage
			err := json.Unmarshal([]byte(ccJSON), &registrationMessage)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(registrationMessage).Should(Equal(CCRegistrationMessage{
				Host: "127.0.0.1",
				Port: 4567,
			}))
		})
	})
})
