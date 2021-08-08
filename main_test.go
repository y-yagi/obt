package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/y-yagi/goext/osext"
)

func TestDownloadTarGz(t *testing.T) {
	setFlags()
	tempDir, err := ioutil.TempDir("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	run([]string{"obt", "-p", "./", "https://github.com/y-yagi/jpcal"}, stdout, stderr)

	if !osext.IsExist("jpcal") {
		t.Fatalf("file download failed")
	}
}

func TestDownloadBinary(t *testing.T) {
	setFlags()
	tempDir, err := ioutil.TempDir("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	run([]string{"obt", "-p", "./", "-tag", "v1.0.0", "https://github.com/davecheney/httpstat"}, stdout, stderr)

	if !osext.IsExist("httpstat") {
		t.Fatalf("file download failed")
	}
}

func TestDownloadRustPackage(t *testing.T) {
	setFlags()
	tempDir, err := ioutil.TempDir("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	run([]string{"obt", "-p", "./", "https://github.com/sharkdp/fd"}, stdout, stderr)

	if !osext.IsExist("fd") {
		t.Fatalf("file download failed")
	}
}
