package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/y-yagi/goext/osext"
)

func TestDownload(t *testing.T) {
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
