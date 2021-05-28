package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
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
	"github.com/ulikunitz/xz"
)

type fileType int

const (
	binary fileType = iota
	tarGzType
	gzipType
	zipType
	tarXzType
)

type downloader struct {
	user       string
	repository string
	url        string
	binaryName string
	fType      fileType
}

func (d *downloader) findDownloadURL() error {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), d.user, d.repository)
	if err != nil {
		return err
	}

	logger.Printf("latest release : %+v\n", *release.TagName)

	for _, asset := range release.Assets {
		if len(d.binaryName) == 0 {
			// TODO(y-yagi): Should I check all assets?
			if a := strings.Split(*asset.Name, "_"); len(a) > 1 {
				d.binaryName = a[0]
			} else {
				d.binaryName = d.repository
			}
		}

		if d.isAvailableBinary(*asset.Name) {
			d.url = *asset.BrowserDownloadURL
			if strings.HasSuffix(*asset.Name, "tar.gz") {
				d.fType = tarGzType
			} else if strings.HasSuffix(*asset.Name, "gzip") {
				d.fType = gzipType
			} else if strings.HasSuffix(*asset.Name, "zip") {
				d.fType = zipType
			} else if strings.HasSuffix(*asset.Name, "tar.xz") {
				d.fType = tarXzType
			} else {
				d.fType = binary
			}

			logger.Printf("download file from : %+v\n", d.url)
			return nil
		}
	}

	msg := fmt.Sprintf("can't find an available released binary. isn't the binary name '%s'?", d.binaryName)
	return errors.New(msg)
}

func (d *downloader) isAvailableBinary(assetName string) bool {
	if !d.isSupportedFormat(assetName) {
		return false
	}

	osAndArch := runtime.GOOS + "_" + runtime.GOARCH

	assetName = strings.Replace(assetName, "-", "_", -1)
	prefix := strings.Replace(d.binaryName, "-", "_", -1)
	assetName = strings.ToLower(assetName)
	if runtime.GOARCH == "amd64" {
		assetName = strings.Replace(assetName, "x86_64", "amd64", -1)
		assetName = strings.Replace(assetName, "64bit", "amd64", -1)
	} else if runtime.GOARCH == "386" {
		assetName = strings.Replace(assetName, "x86", "386", -1)
	}

	return strings.HasPrefix(assetName, prefix) && strings.Contains(assetName, osAndArch)
}

func (d *downloader) execute(file string) error {
	resp, err := http.Get(d.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch d.fType {
	case tarGzType:
		return d.downloadTarGz(&resp.Body, file)
	case gzipType:
		return d.downloadGzip(&resp.Body, file)
	case zipType:
		return d.downloadZip(&resp.Body, file)
	case tarXzType:
		return d.downloadTarXz(&resp.Body, file)
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

		if strings.HasSuffix(hdr.Name, d.binaryName) {
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

func (d *downloader) downloadGzip(body *io.ReadCloser, file string) error {
	r, err := gzip.NewReader(*body)
	if err != nil {
		return err
	}

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, bs, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (d *downloader) downloadZip(body *io.ReadCloser, file string) error {
	zipdata, err := ioutil.ReadAll(*body)
	if err != nil {
		return err
	}

	z, err := zip.NewReader(bytes.NewReader(zipdata), int64(len(zipdata)))
	if err != nil {
		return err
	}

	for _, f := range z.File {
		if strings.HasSuffix(f.Name, d.binaryName) {
			r, err := f.Open()
			if err != nil {
				return err
			}

			b, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(file, b, 0755)
			return err
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

func (d *downloader) downloadTarXz(body *io.ReadCloser, file string) error {
	archive, err := xz.NewReader(*body)
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

		if strings.HasSuffix(hdr.Name, d.binaryName) {
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

func (d *downloader) isSupportedFormat(name string) bool {
	suffixes := []string{"deb", "rpm", "msi", "apk"}
	for _, v := range suffixes {
		if strings.HasSuffix(name, v) {
			return false
		}
	}

	return true
}
