package buildpackrunner

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pivotal-golang/archiver/extractor"
	"github.com/pivotal-golang/cacheddownloader"
)

func IsZipFile(filename string) bool {
	return strings.HasSuffix(filename, ".zip")
}

func DownloadZipAndExtract(u *url.URL, destination string) error {
	zipFile, err := ioutil.TempFile("", filepath.Base(u.Path))
	if err != nil {
		return fmt.Errorf("Could not create zip file: %s", err.Error())
	}
	defer os.Remove(zipFile.Name())

	downloader := cacheddownloader.NewDownloader(DOWNLOAD_TIMEOUT, 1)
	_, _, err = downloader.Download(u,
		func() (*os.File, error) {
			return os.OpenFile(zipFile.Name(), os.O_WRONLY, 0666)
		},
		cacheddownloader.CachingInfoType{},
	)
	if err != nil {
		return fmt.Errorf("Failed to download buildpack '%s': %s", u.String(), err.Error())
	}

	err = extractor.NewZip().Extract(zipFile.Name(), destination)
	if err != nil {
		return fmt.Errorf("Failed to extract buildpack '%s': %s", u.String(), err.Error())
	}

	return nil
}
