package main

import (
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
	run([]string{"obt", "-p", "./", "https://github.com/y-yagi/jpcal"})

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
	run([]string{"obt", "-p", "./", "https://github.com/davecheney/httpstat"})

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
	run([]string{"obt", "-p", "./", "https://github.com/sharkdp/fd"})

	if !osext.IsExist("fd") {
		t.Fatalf("file download failed")
	}
}
