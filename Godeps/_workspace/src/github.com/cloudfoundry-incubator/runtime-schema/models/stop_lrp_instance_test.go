package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StopLrpInstance", func() {
	var stopInstance StopLRPInstance

	stopInstancePayload := `{
		"process_guid":"some-process-guid",
    "instance_guid":"some-instance-guid",
    "index":1234
  }`

	BeforeEach(func() {
		stopInstance = StopLRPInstance{
			ProcessGuid:  "some-process-guid",
			InstanceGuid: "some-instance-guid",
			Index:        1234,
		}
	})
	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json := stopInstance.ToJSON()
			Ω(string(json)).Should(MatchJSON(stopInstancePayload))
		})
	})

	Describe("NewStopLRPInstanceFromJSON", func() {
		It("returns a LRP with correct fields", func() {
			decodedStopInstance, err := NewStopLRPInstanceFromJSON([]byte(stopInstancePayload))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedStopInstance).Should(Equal(stopInstance))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedStopInstance, err := NewStopLRPInstanceFromJSON([]byte("aliens lol"))
				Ω(err).Should(HaveOccurred())

				Ω(decodedStopInstance).Should(BeZero())
			})
		})
	})
})
