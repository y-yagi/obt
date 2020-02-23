package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"github.com/google/go-github/github"
)

type downloader struct {
	user       string
	repository string
	url        string
	fType      fileType
}

func (d *downloader) findDownloadURL() error {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), d.user, d.repository)
	if err != nil {
		return err
	}

	logger.Printf("latest release version: %+v\n", *release.Name)

	for _, asset := range release.Assets {
		if len(binaryName) == 0 {
			// TODO(y-yagi): Should I check all assets?
			if a := strings.Split(*asset.Name, "_"); len(a) > 1 {
				binaryName = a[0]
			} else {
				binaryName = d.repository
			}
		}

		if d.isAvailableBinary(asset) {
			d.url = *asset.BrowserDownloadURL
			if strings.HasSuffix(*asset.Name, "tar.gz") {
				d.fType = tarGz
			} else {
				d.fType = binary
			}
			return nil
		}
	}

	msg := fmt.Sprintf("can't find an available released binary. isn't the binary name '%s'?", binaryName)
	return errors.New(msg)
}

func (d *downloader) isAvailableBinary(asset github.ReleaseAsset) bool {
	if !d.isSupportedFormat(*asset.Name) {
		return false
	}

	target := runtime.GOOS + "_" + runtime.GOARCH

	assetName := strings.Replace(*asset.Name, "-", "_", -1)
	assetName = strings.ToLower(assetName)
	if runtime.GOARCH == "amd64" {
		assetName = strings.Replace(assetName, "x86_64", "amd64", -1)
		assetName = strings.Replace(assetName, "64bit", "amd64", -1)
	} else if runtime.GOARCH == "386" {
		assetName = strings.Replace(assetName, "x86", "386", -1)
	}

	return strings.HasPrefix(assetName, binaryName) && strings.Contains(assetName, target)
}

func (d *downloader) execute(file string) error {
	resp, err := http.Get(d.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if d.fType == tarGz {
		return d.downloadTarGz(&resp.Body, file)
	}

	return d.downloadBinary(&resp.Body, file)
}

func (d *downloader) downloadTarGz(body *io.ReadCloser, file string) error {
	archive, err := gzip.NewReader(*body)
	if err != nil {
		return nil
	}

	tr := tar.NewReader(archive)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil
		}

		if strings.HasSuffix(hdr.Name, binaryName) {
			bs, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil
			}

			err = ioutil.WriteFile(file, bs, 0755)
			if err != nil {
				return nil
			}
			return nil
		}
	}

	return errors.New("can't install released binary. This is a possibility that bug of `obt`. Please report an issue")
}

func (d *downloader) downloadBinary(body *io.ReadCloser, file string) error {
	bs, err := ioutil.ReadAll(*body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, bs, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (d *downloader) isSupportedFormat(name string) bool {
	// TODO(y-yagi): Support zip.
	suffixes := []string{"deb", "rpm", "msi", "zip"}
	for _, v := range suffixes {
		if strings.HasSuffix(name, v) {
			return false
		}
	}

	return true
}
