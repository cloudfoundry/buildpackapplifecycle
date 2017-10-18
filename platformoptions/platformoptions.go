package platformoptions

import (
	"encoding/json"
	"os"
)

type PlatformOptions struct {
	CredhubURI string `json:"credhub-uri"`
}

var cachedPlatformOptions *PlatformOptions

func Get() (*PlatformOptions, error) {
	jsonPlatformOptions := os.Getenv("VCAP_PLATFORM_OPTIONS")
	defer os.Unsetenv("VCAP_PLATFORM_OPTIONS")
	if jsonPlatformOptions != "" {
		platformOptions := PlatformOptions{}
		err := json.Unmarshal([]byte(jsonPlatformOptions), &platformOptions)
		if err != nil {
			return nil, err
		}
		cachedPlatformOptions = &platformOptions
	}
	return cachedPlatformOptions, nil
}
