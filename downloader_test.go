package main

import (
	"io/ioutil"
	"os"
	"runtime"
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

func TestDownloader_Gzip(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	buf, err := ioutil.ReadFile("testdata/sample.gzip")
	if err != nil {
		t.Fatal(err)
	}

	r := ioutil.NopCloser(strings.NewReader(string(buf)))

	downloaded := tempDir + "/sample"
	d := downloader{}
	d.downloadGzip(&r, downloaded)

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

func TestIsAvailableBinary(t *testing.T) {
	osAndArch := runtime.GOOS + "_" + runtime.GOARCH

	var tests = []struct {
		in   string
		want bool
	}{
		{"golangci-lint-1.23.8-" + osAndArch + ".tar.gz", true},
		{"golangci-lint-1.23.8-" + osAndArch + ".deb", false},
		{"golangci-lint-1.23.8-" + osAndArch + ".gzip", true},
	}

	d := downloader{}
	for _, tt := range tests {
		got := d.isAvailableBinary(tt.in)
		if tt.want != got {
			t.Fatalf("in: '%v', expected: %v, got: %v", tt.in, tt.want, got)
		}
	}

}
