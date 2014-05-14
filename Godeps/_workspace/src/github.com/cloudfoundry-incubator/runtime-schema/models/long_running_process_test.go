package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LongRunningProcess", func() {
	var longRunningProcess TransitionalLongRunningProcess

	longRunningProcessPayload := `{
    "guid":"some-guid",
    "stack":"some-stack",
    "actions":[
      {
        "action":"download",
        "args":{
          "from":"old_location",
          "to":"new_location",
          "cache_key":"the-cache-key",
          "extract":true
        }
      }
    ],
    "log": {
      "guid": "123",
      "source_name": "APP",
      "index": 42
    },
    "state": 1
  }`

	BeforeEach(func() {
		index := 42

		longRunningProcess = TransitionalLongRunningProcess{
			Guid:  "some-guid",
			Stack: "some-stack",
			Actions: []ExecutorAction{
				{
					Action: DownloadAction{
						From:     "old_location",
						To:       "new_location",
						CacheKey: "the-cache-key",
						Extract:  true,
					},
				},
			},
			Log: LogConfig{
				Guid:       "123",
				SourceName: "APP",
				Index:      &index,
			},
			State: TransitionalLRPStateDesired,
		}
	})

	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json := longRunningProcess.ToJSON()
			Ω(string(json)).Should(MatchJSON(longRunningProcessPayload))
		})
	})

	Describe("NewLongRunningProcessFromJSON", func() {
		It("returns a LongRunningProcess with correct fields", func() {
			decodedLongRunningProcess, err := NewTransitionalLongRunningProcessFromJSON([]byte(longRunningProcessPayload))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedLongRunningProcess).Should(Equal(longRunningProcess))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedLongRunningProcess, err := NewTransitionalLongRunningProcessFromJSON([]byte("butts lol"))
				Ω(err).Should(HaveOccurred())

				Ω(decodedLongRunningProcess).Should(BeZero())
			})
		})
	})
})
