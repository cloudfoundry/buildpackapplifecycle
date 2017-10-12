package credhub_test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"
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
			server            *ghttp.Server
			fixturesSslDir    string
			userProfile       string
			vcapServices      string
			cfInstanceCert    string
			cfInstanceKey     string
			cfSystemCertsPath string
			err               error
			vcapServicesValue string
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
			userProfile = os.Getenv("USERPROFILE")
			cfInstanceCert = os.Getenv("CF_INSTANCE_CERT")
			cfInstanceKey = os.Getenv("CF_INSTANCE_KEY")
			cfSystemCertsPath = os.Getenv("CF_SYSTEM_CERTS_PATH")
			vcapServices = os.Getenv("VCAP_SERVICES")

			fixturesSslDir, err := filepath.Abs(filepath.Join("..", "fixtures"))
			Expect(err).NotTo(HaveOccurred())

			os.Setenv("USERPROFILE", fixturesSslDir)

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

			if containerpath.For("/") == fixturesSslDir {
				os.Setenv("CF_INSTANCE_CERT", filepath.Join("/certs", "client-tls.crt"))
				os.Setenv("CF_INSTANCE_KEY", filepath.Join("/certs", "client-tls.key"))
				os.Setenv("CF_SYSTEM_CERTS_PATH", "/cacerts")
			} else {
				os.Setenv("CF_INSTANCE_CERT", filepath.Join(fixturesSslDir, "certs", "client-tls.crt"))
				os.Setenv("CF_INSTANCE_KEY", filepath.Join(fixturesSslDir, "certs", "client-tls.key"))
				os.Setenv("CF_SYSTEM_CERTS_PATH", filepath.Join(fixturesSslDir, "cacerts"))
			}
		})

		AfterEach(func() {
			server.Close()
			os.Setenv("USERPROFILE", userProfile)
			os.Setenv("CF_INSTANCE_CERT", cfInstanceCert)
			os.Setenv("CF_INSTANCE_KEY", cfInstanceKey)
			os.Setenv("CF_SYSTEM_CERTS_PATH", cfSystemCertsPath)
			os.Setenv("VCAP_SERVICES", vcapServices)
		})

		BeforeEach(func() {
			vcapServicesValue = `{"my-server":[{"credentials":{"credhub-ref":"(//my-server/creds)"}}]}`
			os.Setenv("VCAP_SERVICES", vcapServicesValue)
		})

		JustBeforeEach(func() {
			err = credhub.InterpolateServiceRefs(server.URL())
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
				Expect(os.Getenv("VCAP_SERVICES")).To(Equal("INTERPOLATED_JSON"))
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
				Expect(os.Getenv("VCAP_SERVICES")).To(Equal(vcapServicesValue))
			})
		})

		Context("when the instance cert and key are invalid", func() {
			BeforeEach(func() {
				if containerpath.For("/") == fixturesSslDir {
					os.Setenv("CF_INSTANCE_CERT", "not_a_cert")
					os.Setenv("CF_INSTANCE_KEY", "not_a_cert")
				} else {
					os.Setenv("CF_INSTANCE_CERT", filepath.Join(fixturesSslDir, "not_a_cert"))
					os.Setenv("CF_INSTANCE_KEY", filepath.Join(fixturesSslDir, "not_a_cert"))
				}
			})

			It("returns an error and doesn't change VCAP_SERVICES", func() {
				Expect(err).To(MatchError(MatchRegexp("Unable to set up credhub client")))
				Expect(os.Getenv("VCAP_SERVICES")).To(Equal(vcapServicesValue))
			})
		})

		Context("when the instance cert and key aren't set", func() {
			BeforeEach(func() {
				os.Unsetenv("CF_INSTANCE_CERT")
				os.Unsetenv("CF_INSTANCE_KEY")
			})

			It("returns an error and doesn't change VCAP_SERVICES", func() {
				Expect(err).To(MatchError(MatchRegexp("Missing CF_INSTANCE_CERT and/or CF_INSTANCE_KEY")))
				Expect(os.Getenv("VCAP_SERVICES")).To(Equal(vcapServicesValue))
			})
		})

		Context("when the system certs path isn't set", func() {
			BeforeEach(func() {
				os.Unsetenv("CF_SYSTEM_CERTS_PATH")
			})

			It("prints an error message", func() {
				Expect(err).To(MatchError(MatchRegexp("Missing CF_SYSTEM_CERTS_PATH")))
				Expect(os.Getenv("VCAP_SERVICES")).To(Equal(vcapServicesValue))
			})
		})
	})
})
