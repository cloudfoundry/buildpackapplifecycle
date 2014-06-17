package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LRPStopAuction", func() {
	var stopAuction LRPStopAuction

	stopAuctionPayload := `{
    "process_guid":"some-guid",
    "index": 2,
    "updated_at": 1138,
    "state": 1
  }`

	BeforeEach(func() {
		stopAuction = LRPStopAuction{
			ProcessGuid: "some-guid",
			Index:       2,
			State:       LRPStopAuctionStatePending,
			UpdatedAt:   1138,
		}
	})
	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json := stopAuction.ToJSON()
			Ω(string(json)).Should(MatchJSON(stopAuctionPayload))
		})
	})

	Describe("NewLRPStopAuctionFromJSON", func() {
		It("returns a LRP with correct fields", func() {
			decodedStopAuction, err := NewLRPStopAuctionFromJSON([]byte(stopAuctionPayload))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedStopAuction).Should(Equal(stopAuction))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedStopAuction, err := NewLRPStopAuctionFromJSON([]byte("aliens lol"))
				Ω(err).Should(HaveOccurred())

				Ω(decodedStopAuction).Should(BeZero())
			})
		})
	})
})
