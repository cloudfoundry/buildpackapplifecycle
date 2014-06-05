package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("RepPresence", func() {
	var repPresence RepPresence

	const payload = `{
    "rep_id":"some-id",
    "stack": "some-stack"
  }`

	BeforeEach(func() {
		repPresence = RepPresence{
			RepID: "some-id",
			Stack: "some-stack",
		}
	})

	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json := repPresence.ToJSON()
			Ω(string(json)).Should(MatchJSON(payload))
		})
	})

	Describe("NewTaskFromJSON", func() {
		It("returns a Task with correct fields", func() {
			decodedRepPresence, err := NewRepPresenceFromJSON([]byte(payload))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedRepPresence).Should(Equal(repPresence))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedRepPresence, err := NewRepPresenceFromJSON([]byte("aliens lol"))
				Ω(err).Should(HaveOccurred())

				Ω(decodedRepPresence).Should(BeZero())
			})
		})
	})
})
