package candiedyaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math"
	"reflect"
	"time"
)

var _ = Describe("Resolver", func() {
	var event yaml_event_t

	var nulls = []string{"", "~", "null", "Null", "NULL"}

	BeforeEach(func() {
		event = yaml_event_t{}
	})

	Context("Implicit events", func() {
		checkNulls := func(f func()) {
			for _, null := range nulls {
				event = yaml_event_t{implicit: true}
				event.value = []byte(null)
				f()
			}
		}

		BeforeEach(func() {
			event.implicit = true
		})

		Context("String", func() {
			It("resolves a string", func() {
				aString := ""
				v := reflect.ValueOf(&aString)
				event.value = []byte("abc")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(aString).To(Equal("abc"))
			})

			It("resolves null", func() {
				checkNulls(func() {
					aString := "abc"
					v := reflect.ValueOf(&aString)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(aString).To(Equal(""))
				})
			})

			It("resolves null pointers", func() {
				checkNulls(func() {
					aString := "abc"
					pString := &aString
					v := reflect.ValueOf(&pString)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(pString).To(BeNil())
				})
			})

		})

		Context("Booleans", func() {
			match_bool := func(val string, expected bool) {
				b := !expected

				v := reflect.ValueOf(&b)
				event.value = []byte(val)

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(b).To(Equal(expected))
			}

			It("resolves on", func() {
				match_bool("on", true)
				match_bool("ON", true)
			})

			It("resolves off", func() {
				match_bool("off", false)
				match_bool("OFF", false)
			})

			It("resolves true", func() {
				match_bool("true", true)
				match_bool("TRUE", true)
			})

			It("resolves false", func() {
				match_bool("false", false)
				match_bool("FALSE", false)
			})

			It("resolves yes", func() {
				match_bool("yes", true)
				match_bool("YES", true)
			})

			It("resolves no", func() {
				match_bool("no", false)
				match_bool("NO", false)
			})

			It("reports an error otherwise", func() {
				b := true
				v := reflect.ValueOf(&b)
				event.value = []byte("fail")

				err := resolve(event, v.Elem())
				Ω(err).Should(HaveOccurred())
			})

			It("resolves null", func() {
				checkNulls(func() {
					b := true
					v := reflect.ValueOf(&b)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(b).To(BeFalse())
				})
			})

			It("resolves null pointers", func() {
				checkNulls(func() {
					b := true
					pb := &b
					v := reflect.ValueOf(&pb)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(pb).To(BeNil())
				})
			})
		})

		Context("Ints", func() {
			It("simple ints", func() {
				i := 0
				v := reflect.ValueOf(&i)
				event.value = []byte("1234")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(1234))
			})

			It("positive ints", func() {
				i := int16(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("+678")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(int16(678)))
			})

			It("negative ints", func() {
				i := int32(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("-2345")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(int32(-2345)))
			})

			It("base 2", func() {
				i := 0
				v := reflect.ValueOf(&i)
				event.value = []byte("0b11")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(3))
			})

			It("base 8", func() {
				i := 0
				v := reflect.ValueOf(&i)
				event.value = []byte("012")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(10))
			})

			It("base 16", func() {
				i := 0
				v := reflect.ValueOf(&i)
				event.value = []byte("0xff")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(255))
			})

			It("base 60", func() {
				i := 0
				v := reflect.ValueOf(&i)
				event.value = []byte("1:30:00")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(5400))
			})

			It("fails on overflow", func() {
				i := int8(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("2345")

				err := resolve(event, v.Elem())
				Ω(err).Should(HaveOccurred())
			})

			It("fails on invalid int", func() {
				i := 0
				v := reflect.ValueOf(&i)
				event.value = []byte("234f")

				err := resolve(event, v.Elem())
				Ω(err).Should(HaveOccurred())
			})

			It("resolves null", func() {
				checkNulls(func() {
					i := 1
					v := reflect.ValueOf(&i)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(i).To(Equal(0))
				})
			})

			It("resolves null pointers", func() {
				checkNulls(func() {
					i := 1
					pi := &i
					v := reflect.ValueOf(&pi)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(pi).To(BeNil())
				})
			})
		})

		Context("UInts", func() {
			It("resolves simple uints", func() {
				i := uint(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("1234")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(uint(1234)))
			})

			It("resolves positive uints", func() {
				i := uint16(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("+678")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(uint16(678)))
			})

			It("base 2", func() {
				i := uint(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("0b11")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(uint(3)))
			})

			It("base 8", func() {
				i := uint(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("012")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(uint(10)))
			})

			It("base 16", func() {
				i := uint(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("0xff")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(uint(255)))
			})

			It("base 60", func() {
				i := uint(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("1:30:01")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(i).To(Equal(uint(5401)))
			})

			It("fails with negative ints", func() {
				i := uint(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("-2345")

				err := resolve(event, v.Elem())
				Ω(err).Should(HaveOccurred())
			})

			It("fails on overflow", func() {
				i := uint8(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("2345")

				err := resolve(event, v.Elem())
				Ω(err).Should(HaveOccurred())
			})

			It("resolves null", func() {
				checkNulls(func() {
					i := uint(1)
					v := reflect.ValueOf(&i)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(i).To(Equal(uint(0)))
				})
			})

			It("resolves null pointers", func() {
				checkNulls(func() {
					i := uint(1)
					pi := &i
					v := reflect.ValueOf(&pi)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(pi).To(BeNil())
				})
			})

		})

		Context("Floats", func() {
			It("float32", func() {
				f := float32(0)
				v := reflect.ValueOf(&f)
				event.value = []byte("2345.01")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(f).To(Equal(float32(2345.01)))
			})

			It("float64", func() {
				f := float64(0)
				v := reflect.ValueOf(&f)
				event.value = []byte("-456456.01")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(f).To(Equal(float64(-456456.01)))
			})

			It("+inf", func() {
				f := float64(0)
				v := reflect.ValueOf(&f)
				event.value = []byte("+.inf")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(f).To(Equal(math.Inf(1)))
			})

			It("-inf", func() {
				f := float32(0)
				v := reflect.ValueOf(&f)
				event.value = []byte("-.inf")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(f).To(Equal(float32(math.Inf(-1))))
			})

			It("nan", func() {
				f := float64(0)
				v := reflect.ValueOf(&f)
				event.value = []byte(".NaN")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(math.IsNaN(f)).To(BeTrue())
			})

			It("base 60", func() {
				f := float64(0)
				v := reflect.ValueOf(&f)
				event.value = []byte("1:30:02")

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(f).To(Equal(float64(5402)))
			})

			It("fails on overflow", func() {
				i := float32(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("123e10000")

				err := resolve(event, v.Elem())
				Ω(err).Should(HaveOccurred())
			})

			It("fails on invalid float", func() {
				i := float32(0)
				v := reflect.ValueOf(&i)
				event.value = []byte("123e1a")

				err := resolve(event, v.Elem())
				Ω(err).Should(HaveOccurred())
			})

			It("resolves null", func() {
				checkNulls(func() {
					f := float64(1)
					v := reflect.ValueOf(&f)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(f).To(Equal(0.0))
				})
			})

			It("resolves null pointers", func() {
				checkNulls(func() {
					f := float64(1)
					pf := &f
					v := reflect.ValueOf(&pf)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(pf).To(BeNil())
				})
			})
		})

		Context("Timestamps", func() {
			parse_date := func(val string, date time.Time) {
				d := time.Now()
				v := reflect.ValueOf(&d)
				event.value = []byte(val)

				err := resolve(event, v.Elem())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(d).To(Equal(date))
			}

			It("date", func() {
				parse_date("2002-12-14", time.Date(2002, time.December, 14, 0, 0, 0, 0, time.UTC))
			})

			It("canonical", func() {
				parse_date("2001-12-15T02:59:43.1Z", time.Date(2001, time.December, 15, 2, 59, 43, int(1*time.Millisecond), time.UTC))
			})

			It("iso8601", func() {
				parse_date("2001-12-14t21:59:43.10-05:00", time.Date(2001, time.December, 14, 21, 59, 43, int(10*time.Millisecond), time.FixedZone("", -5*3600)))
			})

			It("space separated", func() {
				parse_date("2001-12-14 21:59:43.10 -5", time.Date(2001, time.December, 14, 21, 59, 43, int(10*time.Millisecond), time.FixedZone("", -5*3600)))
			})

			It("no time zone", func() {
				parse_date("2001-12-15 2:59:43.10", time.Date(2001, time.December, 15, 2, 59, 43, int(10*time.Millisecond), time.UTC))
			})

			It("resolves null", func() {
				checkNulls(func() {
					d := time.Now()
					v := reflect.ValueOf(&d)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(d).To(Equal(time.Time{}))
				})
			})

			It("resolves null pointers", func() {
				checkNulls(func() {
					d := time.Now()
					pd := &d
					v := reflect.ValueOf(&pd)

					err := resolve(event, v.Elem())
					Ω(err).ShouldNot(HaveOccurred())
					Ω(pd).To(BeNil())
				})
			})
		})

		It("fails to resolve a pointer", func() {
			aString := ""
			pString := &aString
			v := reflect.ValueOf(&pString)
			event.value = []byte("abc")

			err := resolve(event, v.Elem())
			Ω(err).Should(HaveOccurred())
		})

	})
})
