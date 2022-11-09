package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/y-yagi/goext/osext"
)

func TestDownloadTarGz(t *testing.T) {
	setFlags()
	origiDir, _ := os.Getwd()
	tempDir, err := os.MkdirTemp("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	defer os.Chdir(origiDir)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	run([]string{"obt", "-p", "./", "https://github.com/y-yagi/jpcal"}, stdout, stderr)

	if !osext.IsExist("jpcal") {
		t.Fatalf("file download failed")
	}
}

func TestDownloadBinary(t *testing.T) {
	setFlags()
	origiDir, _ := os.Getwd()
	tempDir, err := os.MkdirTemp("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	defer os.Chdir(origiDir)

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	run([]string{"obt", "-p", "./", "-tag", "v1.0.0", "https://github.com/davecheney/httpstat"}, stdout, stderr)

	if !osext.IsExist("httpstat") {
		t.Fatalf("file download failed")
	}
}

func TestDownloadRustPackage(t *testing.T) {
	setFlags()
	origiDir, _ := os.Getwd()
	tempDir, err := os.MkdirTemp("", "obttest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	defer os.Chdir(origiDir)

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	run([]string{"obt", "-p", "./", "https://github.com/sharkdp/fd"}, stdout, stderr)

	if !osext.IsExist("fd") {
		t.Fatalf("file download failed")
	}
}

func TestShowHistory(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	abs, _ := filepath.Abs("./testdata/")
	run([]string{"obt", "-history", abs}, stdout, stderr)

	stdout, stderr = new(bytes.Buffer), new(bytes.Buffer)
	run([]string{"obt", "-installed"}, stdout, stderr)

	want := `
+---------------------------------+--------+--------------------------+
|               URL               |  TAG   |           PATH           |
+---------------------------------+--------+--------------------------+
| https://github.com/y-yagi/jpcal | v1.0.2 | /home/y-yagi/gobin/jpcal |
+---------------------------------+--------+--------------------------+
`

	got := "\n" + stdout.String()
	if string(got) != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestSetInvalidPathToDefaultPath(t *testing.T) {
	setFlags()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	run([]string{"obt", "-s", "a"}, stdout, stderr)

	want := "Please specify an absolute path to the default install path.\n"
	got := stderr.String()
	if string(got) != want {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}
