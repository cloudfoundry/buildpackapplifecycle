package credhub_test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"code.cloudfoundry.org/buildpackapplifecycle/containerpath"
	"code.cloudfoundry.org/buildpackapplifecycle/credhub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("credhub", func() {
	Describe("InterpolateServiceRefs", func() {
		var (
			fakeEnv           map[string]string
			vcapServicesValue string
			server            *ghttp.Server
			fixturesSslDir    string
			err               error

			subject *credhub.Credhub
		)

		VerifyClientCerts := func() http.HandlerFunc {
			return func(w http.ResponseWriter, req *http.Request) {
				tlsConnectionState := req.TLS
				Expect(tlsConnectionState).NotTo(BeNil())
				Expect(tlsConnectionState.PeerCertificates).NotTo(BeEmpty())
				Expect(tlsConnectionState.PeerCertificates[0].Subject.CommonName).To(Equal("example.com"))
			}
		}

		BeforeEach(func() {
			fakeEnv = make(map[string]string)

			fixturesSslDir, err := filepath.Abs(filepath.Join("..", "fixtures"))
			Expect(err).NotTo(HaveOccurred())

			server = ghttp.NewUnstartedServer()

			cert, err := tls.LoadX509KeyPair(filepath.Join(fixturesSslDir, "certs", "server-tls.crt"), filepath.Join(fixturesSslDir, "certs", "server-tls.key"))
			Expect(err).NotTo(HaveOccurred())

			caCerts := x509.NewCertPool()

			caCertBytes, err := ioutil.ReadFile(filepath.Join(fixturesSslDir, "cacerts", "client-tls-ca.crt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(caCerts.AppendCertsFromPEM(caCertBytes)).To(BeTrue())

			server.HTTPTestServer.TLS = &tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{cert},
				ClientCAs:    caCerts,
			}
			server.HTTPTestServer.StartTLS()

			cpath := containerpath.New(fixturesSslDir)
			if cpath.For("/") == fixturesSslDir {
				fakeEnv["CF_INSTANCE_CERT"] = filepath.Join("/certs", "client-tls.crt")
				fakeEnv["CF_INSTANCE_KEY"] = filepath.Join("/certs", "client-tls.key")
				fakeEnv["CF_SYSTEM_CERT_PATH"] = "/cacerts"
			} else {
				fakeEnv["CF_INSTANCE_CERT"] = filepath.Join(fixturesSslDir, "certs", "client-tls.crt")
				fakeEnv["CF_INSTANCE_KEY"] = filepath.Join(fixturesSslDir, "certs", "client-tls.key")
				fakeEnv["CF_SYSTEM_CERT_PATH"] = filepath.Join(fixturesSslDir, "cacerts")
			}

			subject = &credhub.Credhub{
				Setenv: func(key, value string) error {
					fakeEnv[key] = value
					return nil
				},
				Getenv: func(key string) string {
					return fakeEnv[key]
				},
				PathFor: containerpath.New(fixturesSslDir).For,
			}
		})

		AfterEach(func() {
			server.Close()
		})

		BeforeEach(func() {
			vcapServicesValue = `{"my-server":[{"credentials":{"credhub-ref":"(//my-server/creds)"}}]}`
			fakeEnv["VCAP_SERVICES"] = vcapServicesValue
		})

		JustBeforeEach(func() {
			err = subject.InterpolateServiceRefs(server.URL())
		})

		Context("when there are no credhub refs in VCAP_SERVICES and no TLS environment variables are present", func() {
			BeforeEach(func() {
				delete(fakeEnv, "CF_INSTANCE_CERT")
				delete(fakeEnv, "CF_INSTANCE_KEY")
				delete(fakeEnv, "CF_SYSTEM_CERT_PATH")

				vcapServicesValue = `{"my-server":[{"credentials":{"no refs here":"and this string containing credhub-ref doesnt count"}}]}`
				fakeEnv["VCAP_SERVICES"] = vcapServicesValue
			})

			It("does not fail and does not change VCAP_SERVICES", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeEnv["VCAP_SERVICES"]).To(Equal(vcapServicesValue))
			})
		})

		Context("when credhub successfully interpolates", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v1/interpolate"),
						ghttp.VerifyBody([]byte(vcapServicesValue)),
						VerifyClientCerts(),
						ghttp.RespondWith(http.StatusOK, "INTERPOLATED_JSON"),
					))
			})

			It("updates VCAP_SERVICES with the interpolated content and runs the process without VCAP_PLATFORM_OPTIONS", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeEnv["VCAP_SERVICES"]).To(Equal("INTERPOLATED_JSON"))
			})
		})

		Context("when credhub fails", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v1/interpolate"),
						ghttp.VerifyBody([]byte(vcapServicesValue)),
						ghttp.RespondWith(http.StatusInternalServerError, "{}"),
					))
			})

			It("returns an error and doesn't change VCAP_SERVICES", func() {
				Expect(err).To(MatchError(MatchRegexp("Unable to interpolate credhub references")))
				Expect(fakeEnv["VCAP_SERVICES"]).To(Equal(vcapServicesValue))
			})
		})

		Context("when the instance cert and key are invalid", func() {
			BeforeEach(func() {
				cpath := containerpath.New(fixturesSslDir)
				if cpath.For("/") == fixturesSslDir {
					fakeEnv["CF_INSTANCE_CERT"] = "not_a_cert"
					fakeEnv["CF_INSTANCE_KEY"] = "not_a_cert"
				} else {
					fakeEnv["CF_INSTANCE_CERT"] = filepath.Join(fixturesSslDir, "not_a_cert")
					fakeEnv["CF_INSTANCE_KEY"] = filepath.Join(fixturesSslDir, "not_a_cert")
				}
			})

			It("returns an error and doesn't change VCAP_SERVICES", func() {
				Expect(err).To(MatchError(MatchRegexp("Unable to set up credhub client")))
				Expect(fakeEnv["VCAP_SERVICES"]).To(Equal(vcapServicesValue))
			})
		})

		Context("when the instance cert and key aren't set", func() {
			BeforeEach(func() {
				delete(fakeEnv, "CF_INSTANCE_CERT")
				delete(fakeEnv, "CF_INSTANCE_KEY")
			})

			It("returns an error and doesn't change VCAP_SERVICES", func() {
				Expect(err).To(MatchError(MatchRegexp("Missing CF_INSTANCE_CERT and/or CF_INSTANCE_KEY")))
				Expect(fakeEnv["VCAP_SERVICES"]).To(Equal(vcapServicesValue))
			})
		})

		Context("when the system certs path isn't set", func() {
			BeforeEach(func() {
				delete(fakeEnv, "CF_SYSTEM_CERT_PATH")
			})

			It("prints an error message", func() {
				Expect(err).To(MatchError(MatchRegexp("Missing CF_SYSTEM_CERT_PATH")))
				Expect(fakeEnv["VCAP_SERVICES"]).To(Equal(vcapServicesValue))
			})
		})
	})
})
