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
	if jsonPlatformOptions != "" {
		platformOptions := PlatformOptions{}
		err := json.Unmarshal([]byte(jsonPlatformOptions), &platformOptions)
		if err != nil {
			return nil, err
		}
		return &platformOptions, nil
	}
	return nil, nil
}
