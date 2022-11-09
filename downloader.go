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
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
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

type Downloader struct {
	user       string
	repository string
	url        string
	binaryName string
	fType      fileType
	cachePath  string
	releaseTag string
}

func (d *Downloader) findDownloadURL() error {
	var client *github.Client

	if len(d.cachePath) != 0 {
		client = github.NewClient(httpcache.NewTransport(diskcache.New(d.cachePath)).Client())
		logger.Printf("use httpcache. path: %+v\n", d.cachePath)
	} else {
		client = github.NewClient(nil)
	}

	var release *github.RepositoryRelease
	var err error

	if len(d.releaseTag) != 0 {
		release, _, err = client.Repositories.GetReleaseByTag(context.Background(), d.user, d.repository, d.releaseTag)
		if err != nil {
			return err
		}
	} else {
		release, _, err = client.Repositories.GetLatestRelease(context.Background(), d.user, d.repository)
		if err != nil {
			return err
		}

		logger.Printf("latest release : %+v\n", *release.TagName)
		d.releaseTag = *release.TagName
	}

	for _, asset := range release.Assets {
		if len(d.binaryName) == 0 {
			if strings.Contains(*asset.Name, d.repository) {
				d.binaryName = d.repository
			} else if a := strings.Split(*asset.Name, "_"); len(a) > 1 {
				// TODO(y-yagi): Should I check all assets?
				d.binaryName = a[0]
			} else {
				d.binaryName = d.repository
			}
		}

		if d.isAvailableBinary(*asset.Name) {
			d.url = *asset.BrowserDownloadURL
			switch {
			case strings.HasSuffix(*asset.Name, "tar.gz"):
				d.fType = tarGzType
			case strings.HasSuffix(*asset.Name, "gzip"):
				d.fType = gzipType
			case strings.HasSuffix(*asset.Name, "zip"):
				d.fType = zipType
			case strings.HasSuffix(*asset.Name, "tar.xz"):
				d.fType = tarXzType
			case strings.HasSuffix(*asset.Name, "gz"):
				d.fType = gzipType
			default:
				d.fType = binary
			}

			logger.Printf("download file from : %+v\n", d.url)
			return nil
		}
	}

	msg := fmt.Sprintf("can't find an available released binary. isn't the binary name '%s'?", d.binaryName)
	return errors.New(msg)
}

func (d *Downloader) isAvailableBinary(assetName string) bool {
	if !d.isSupportedFormat(assetName) {
		return false
	}

	assetName = strings.Replace(assetName, "-", "_", -1)
	prefix := strings.Replace(d.binaryName, "-", "_", -1)
	assetName = strings.ToLower(assetName)
	if runtime.GOARCH == "amd64" {
		assetName = strings.Replace(assetName, "x86_64", "amd64", -1)
		assetName = strings.Replace(assetName, "64bit", "amd64", -1)
	} else if runtime.GOARCH == "386" {
		assetName = strings.Replace(assetName, "x86", "386", -1)
	}

	return strings.HasPrefix(assetName, prefix) && strings.Contains(assetName, runtime.GOOS) && strings.Contains(assetName, runtime.GOARCH)
}

func (d *Downloader) execute(file string) error {
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

func (d *Downloader) downloadTarGz(body *io.ReadCloser, file string) error {
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

		if filepath.Base(hdr.Name) == d.binaryName {
			bs, err := io.ReadAll(tr)
			if err != nil {
				return nil
			}

			return d.writeFile(file, bs)
		}
	}

	return errors.New("can't install released binary. This is a possibility that bug of `obt`. Please report an issue")
}

func (d *Downloader) downloadGzip(body *io.ReadCloser, file string) error {
	r, err := gzip.NewReader(*body)
	if err != nil {
		return err
	}

	bs, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return d.writeFile(file, bs)
}

func (d *Downloader) downloadZip(body *io.ReadCloser, file string) error {
	zipdata, err := io.ReadAll(*body)
	if err != nil {
		return err
	}

	z, err := zip.NewReader(bytes.NewReader(zipdata), int64(len(zipdata)))
	if err != nil {
		return err
	}

	for _, f := range z.File {
		if filepath.Base(f.Name) == d.binaryName {
			r, err := f.Open()
			if err != nil {
				return err
			}

			b, err := io.ReadAll(r)
			if err != nil {
				return err
			}

			return d.writeFile(file, b)
		}
	}
	return errors.New("can't install released binary. This is a possibility that bug of `obt`. Please report an issue")
}

func (d *Downloader) downloadBinary(body *io.ReadCloser, file string) error {
	bs, err := io.ReadAll(*body)
	if err != nil {
		return err
	}

	return d.writeFile(file, bs)
}

func (d *Downloader) downloadTarXz(body *io.ReadCloser, file string) error {
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

		if filepath.Base(hdr.Name) == d.binaryName {
			bs, err := io.ReadAll(tr)
			if err != nil {
				return nil
			}

			return d.writeFile(file, bs)
		}
	}

	return errors.New("can't install released binary. This is a possibility that bug of `obt`. Please report an issue")
}

func (d *Downloader) isSupportedFormat(name string) bool {
	suffixes := []string{"deb", "rpm", "msi", "apk"}
	for _, v := range suffixes {
		if strings.HasSuffix(name, v) {
			return false
		}
	}

	return true
}

func (d *Downloader) writeFile(file string, b []byte) error {
	return os.WriteFile(file, b, 0755)
}
