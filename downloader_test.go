package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/y-yagi/goext/osext"
)

func TestDownloader_TarGz(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	buf, err := ioutil.ReadFile("testdata/sample.tar.gz")
	if err != nil {
		t.Fatal(err)
	}

	r := ioutil.NopCloser(strings.NewReader(string(buf)))

	downloaded := tempDir + "/sample"
	d := downloader{}
	d.downloadTarGz(&r, downloaded)

	if !osext.IsExist(downloaded) {
		t.Fatalf("file download failed")
	}

	buf, err = ioutil.ReadFile(downloaded)
	if err != nil {
		t.Fatal(err)
	}

	want := "sample\n"
	if string(buf) != want {
		t.Fatalf("expected '%s', but got '%s'\n", want, buf)
	}
}

func TestDownloader_Zip(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	buf, err := ioutil.ReadFile("testdata/sample.zip")
	if err != nil {
		t.Fatal(err)
	}

	r := ioutil.NopCloser(strings.NewReader(string(buf)))

	downloaded := tempDir + "/sample"
	d := downloader{}
	d.downloadZip(&r, downloaded)

	if !osext.IsExist(downloaded) {
		t.Fatalf("file download failed")
	}

	buf, err = ioutil.ReadFile(downloaded)
	if err != nil {
		t.Fatal(err)
	}

	want := "sample\n"
	if string(buf) != want {
		t.Fatalf("expected '%s', but got '%s'\n", want, buf)
	}
}
