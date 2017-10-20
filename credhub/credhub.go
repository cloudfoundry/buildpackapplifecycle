package credhub

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/buildpackapplifecycle/containerpath"
	api "github.com/cloudfoundry-incubator/credhub-cli/credhub"
)

type Credhub struct {
	Setenv func(key, value string) error
	Getenv func(key string) string
}

func New() *Credhub {
	return &Credhub{
		Setenv: os.Setenv,
		Getenv: os.Getenv,
	}
}

func (c *Credhub) InterpolateServiceRefs(credhubURI string) error {
	if !strings.Contains(c.Getenv("VCAP_SERVICES"), `"credhub-ref"`) {
		return nil
	}
	ch, err := c.credhubClient(credhubURI)
	if err != nil {
		return fmt.Errorf("Unable to set up credhub client: %v", err)
	}
	interpolatedServices, err := ch.InterpolateString(c.Getenv("VCAP_SERVICES"))
	if err != nil {
		return fmt.Errorf("Unable to interpolate credhub references: %v", err)
	}
	c.Setenv("VCAP_SERVICES", interpolatedServices)
	return nil
}

func (c *Credhub) credhubClient(credhubURI string) (*api.CredHub, error) {
	if c.Getenv("CF_INSTANCE_CERT") == "" || c.Getenv("CF_INSTANCE_KEY") == "" {
		return nil, fmt.Errorf("Missing CF_INSTANCE_CERT and/or CF_INSTANCE_KEY")
	}
	if c.Getenv("CF_SYSTEM_CERT_PATH") == "" {
		return nil, fmt.Errorf("Missing CF_SYSTEM_CERT_PATH")
	}

	systemCertsPath := containerpath.For(c.Getenv("CF_SYSTEM_CERT_PATH"))
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

	return api.New(
		credhubURI,
		api.ClientCert(containerpath.For(c.Getenv("CF_INSTANCE_CERT")), containerpath.For(c.Getenv("CF_INSTANCE_KEY"))),
		api.CaCerts(caCerts...),
	)
}
