package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/github"
)

const cmd = "obt"

var (
	flags       *flag.FlagSet
	showVersion bool
	path        string
	binaryName  string

	version = "devel"
)

func init() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.BoolVar(&showVersion, "v", false, "print version number")
	flags.StringVar(&path, "p", "", "install path")
	flags.StringVar(&binaryName, "b", "", "binary name(default: repository name)")
	flags.Usage = usage
}

func main() {
	os.Exit(run(os.Args))
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n\n", cmd)
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flags.PrintDefaults()
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmd, err)
		return 1
	}
	return 0
}

func run(args []string) int {
	flags.Parse(args[1:])

	if showVersion {
		fmt.Fprintf(os.Stdout, "%s %s (runtime: %s)\n", cmd, version, runtime.Version())
		return 0
	}

	if len(flags.Args()) == 0 {
		flags.Usage()
		return 0
	}

	a := strings.Split(flags.Args()[0], "/")
	userName := a[len(a)-2]
	repo := a[len(a)-1]

	if len(binaryName) == 0 {
		binaryName = repo
	}

	url, err := findDownloadURL(userName, repo)
	if err != nil {
		return msg(err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return msg(err)
	}

	defer resp.Body.Close()

	archive, err := gzip.NewReader(resp.Body)
	if err != nil {
		return msg(err)
	}

	tr := tar.NewReader(archive)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return msg(err)
		}

		if strings.HasSuffix(hdr.Name, binaryName) {
			bs, err := ioutil.ReadAll(tr)
			if err != nil {
				return msg(err)
			}

			file := filepath.Join(strings.TrimSuffix(determinePath(), "\n"), binaryName)
			err = ioutil.WriteFile(file, bs, 0755)
			if err != nil {
				return msg(err)
			}

			fmt.Fprintf(os.Stdout, "Install '%s' to '%s'.\n", binaryName, file)
			return 0
		}
	}

	return msg(errors.New("can't install released binary"))
}

func determinePath() string {
	if len(path) > 0 {
		return path
	}

	gobin, err := exec.Command("go", "env", "GOBIN").Output()
	if err == nil && len(gobin) > 0 {
		return string(gobin)
	}

	// TODO: support windows
	return "/usr/local/bin/"
}

func findDownloadURL(userName, repo string) (string, error) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), userName, repo)
	if err != nil {
		return "", err
	}

	suffix := runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	for _, asset := range release.Assets {
		name := strings.Replace(*asset.Name, "-", "_", -1)
		if strings.HasPrefix(name, binaryName) && strings.HasSuffix(name, suffix) {
			return *asset.BrowserDownloadURL, nil
		}
	}

	return "", errors.New("can't find released binary")
}
