package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LRP", func() {
	Describe("LRPStartAuction", func() {
		var startAuction LRPStartAuction

		startAuctionPayload := `{
    "process_guid":"some-guid",
    "instance_guid":"some-instance-guid",
    "stack":"some-stack",
    "memory_mb" : 128,
    "disk_mb" : 512,
    "ports": [
      { "container_port": 8080 },
      { "container_port": 8081, "host_port": 1234 }
    ],
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
    "index": 2,
    "state": 1
  }`

		BeforeEach(func() {
			index := 42

			startAuction = LRPStartAuction{
				ProcessGuid:  "some-guid",
				InstanceGuid: "some-instance-guid",
				Stack:        "some-stack",
				MemoryMB:     128,
				DiskMB:       512,
				Ports: []PortMapping{
					{ContainerPort: 8080},
					{ContainerPort: 8081, HostPort: 1234},
				},
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
				Index: 2,
				State: LRPStartAuctionStatePending,
			}
		})

		Describe("ToJSON", func() {
			It("should JSONify", func() {
				json := startAuction.ToJSON()
				Ω(string(json)).Should(MatchJSON(startAuctionPayload))
			})
		})

		Describe("NewLRPStartAuctionFromJSON", func() {
			It("returns a LRP with correct fields", func() {
				decodedStartAuction, err := NewLRPStartAuctionFromJSON([]byte(startAuctionPayload))
				Ω(err).ShouldNot(HaveOccurred())

				Ω(decodedStartAuction).Should(Equal(startAuction))
			})

			Context("with an invalid payload", func() {
				It("returns the error", func() {
					decodedStartAuction, err := NewLRPStartAuctionFromJSON([]byte("butts lol"))
					Ω(err).Should(HaveOccurred())

					Ω(decodedStartAuction).Should(BeZero())
				})
			})
		})
	})

	Describe("LRP", func() {
		var lrp LRP

		lrpPayload := `{
    "process_guid":"some-guid",
    "instance_guid":"some-instance-guid",
		"host": "1.2.3.4",
    "ports": [
      { "container_port": 8080 },
      { "container_port": 8081, "host_port": 1234 }
    ],
    "index": 2,
    "state": 0
  }`

		BeforeEach(func() {
			lrp = LRP{
				ProcessGuid:  "some-guid",
				InstanceGuid: "some-instance-guid",
				Host:         "1.2.3.4",
				Ports: []PortMapping{
					{ContainerPort: 8080},
					{ContainerPort: 8081, HostPort: 1234},
				},
				Index: 2,
			}
		})

		Describe("ToJSON", func() {
			It("should JSONify", func() {
				json := lrp.ToJSON()
				Ω(string(json)).Should(MatchJSON(lrpPayload))
			})
		})

		Describe("NewLRPFromJSON", func() {
			It("returns a LRP with correct fields", func() {
				decodedStartAuction, err := NewLRPFromJSON([]byte(lrpPayload))
				Ω(err).ShouldNot(HaveOccurred())

				Ω(decodedStartAuction).Should(Equal(lrp))
			})

			Context("with an invalid payload", func() {
				It("returns the error", func() {
					decodedStartAuction, err := NewLRPFromJSON([]byte("butts lol"))
					Ω(err).Should(HaveOccurred())

					Ω(decodedStartAuction).Should(BeZero())
				})
			})
		})
	})

	Describe("DesiredLRP", func() {
		var lrp DesiredLRP

		lrpPayload := `{
    "process_guid":"some-guid",
    "instances":5,
		"stack":"some-stack",
		"memory_mb":1024,
		"disk_mb":512,
		"routes":["route-1","route-2"]
  }`

		BeforeEach(func() {
			lrp = DesiredLRP{
				ProcessGuid: "some-guid",
				Instances:   5,
				Stack:       "some-stack",
				MemoryMB:    1024,
				DiskMB:      512,
				Routes:      []string{"route-1", "route-2"},
			}
		})

		Describe("ToJSON", func() {
			It("should JSONify", func() {
				json := lrp.ToJSON()
				Ω(string(json)).Should(MatchJSON(lrpPayload))
			})
		})

		Describe("NewDesiredLRPFromJSON", func() {
			It("returns a LRP with correct fields", func() {
				decodedStartAuction, err := NewDesiredLRPFromJSON([]byte(lrpPayload))
				Ω(err).ShouldNot(HaveOccurred())

				Ω(decodedStartAuction).Should(Equal(lrp))
			})

			Context("with an invalid payload", func() {
				It("returns the error", func() {
					decodedStartAuction, err := NewDesiredLRPFromJSON([]byte("butts lol"))
					Ω(err).Should(HaveOccurred())

					Ω(decodedStartAuction).Should(BeZero())
				})
			})
		})
	})
})
