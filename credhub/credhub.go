package credhub

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/buildpackapplifecycle/containerpath"
	"github.com/cloudfoundry-incubator/credhub-cli/credhub"
)

func InterpolateServiceRefs(credhubURI string) error {
	ch, err := credhubClient(credhubURI)
	if err != nil {
		return fmt.Errorf("Unable to set up credhub client: %v", err)
	}
	interpolatedServices, err := ch.InterpolateString(os.Getenv("VCAP_SERVICES"))
	if err != nil {
		return fmt.Errorf("Unable to interpolate credhub references: %v", err)
	}
	os.Setenv("VCAP_SERVICES", interpolatedServices)
	return nil
}

func credhubClient(credhubURI string) (*credhub.CredHub, error) {
	if os.Getenv("CF_INSTANCE_CERT") == "" || os.Getenv("CF_INSTANCE_KEY") == "" {
		return nil, fmt.Errorf("Missing CF_INSTANCE_CERT and/or CF_INSTANCE_KEY")
	}
	if os.Getenv("CF_SYSTEM_CERTS_PATH") == "" {
		return nil, fmt.Errorf("Missing CF_SYSTEM_CERTS_PATH")
	}

	systemCertsPath := containerpath.For(os.Getenv("CF_SYSTEM_CERTS_PATH"))
	caCerts := []string{}
	files, err := ioutil.ReadDir(systemCertsPath)
	if err != nil {
		return nil, fmt.Errorf("Can't read contents of system cert path: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".crt") {
			contents, err := ioutil.ReadFile(filepath.Join(systemCertsPath, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("Can't read contents of cert in system cert path: %v", err)
			}
			caCerts = append(caCerts, string(contents))
		}
	}

	return credhub.New(
		credhubURI,
		credhub.ClientCert(containerpath.For(os.Getenv("CF_INSTANCE_CERT")), containerpath.For(os.Getenv("CF_INSTANCE_KEY"))),
		credhub.CaCerts(caCerts...),
	)
}
