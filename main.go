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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/github"
	"github.com/y-yagi/configure"
	"github.com/y-yagi/debuglog"
)

const cmd = "obt"

var (
	cfg    config
	logger *debuglog.Logger

	flags       *flag.FlagSet
	showVersion bool
	path        string
	defaultPath string
	binaryName  string

	version = "devel"
)

type config struct {
	Path string `toml:"path"`
}

type fileType int

const (
	unknown fileType = iota
	binary
	tarGz
)

func main() {
	setFlags()
	os.Exit(run(os.Args))
}

func setFlags() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.BoolVar(&showVersion, "v", false, "print version number")
	flags.StringVar(&path, "p", "", "install path")
	flags.StringVar(&defaultPath, "s", "", "set default install path")
	flags.StringVar(&binaryName, "b", "", "binary name(default: repository name)")
	flags.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n\n", cmd)
	fmt.Fprintf(os.Stderr, "Install binary file from GitHub's release page. Default install path is '%s'.\n\n", determinePath())
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
	logger = debuglog.New(os.Stdout)
	configure.Load(cmd, &cfg)

	flags.Parse(args[1:])

	if showVersion {
		fmt.Fprintf(os.Stdout, "%s %s (runtime: %s)\n", cmd, version, runtime.Version())
		return 0
	}

	if len(defaultPath) > 0 {
		cfg.Path = defaultPath
		configure.Save(cmd, cfg)
		fmt.Fprintf(os.Stdout, "Change default install path to '%s'\n", defaultPath)
		return 0
	}

	if len(flags.Args()) == 0 {
		flags.Usage()
		return 0
	}

	url := strings.TrimSuffix(flags.Args()[0], "/")
	a := strings.Split(url, "/")
	userName := a[len(a)-2]
	repo := a[len(a)-1]

	if len(binaryName) == 0 {
		binaryName = repo
	}

	url, ft, err := findDownloadURL(userName, repo)
	if err != nil {
		return msg(err)
	}

	logger.Printf("download file from '%+v'\n", url)
	path := determinePath()
	if _, err := os.Stat(path); err != nil {
		return msg(err)
	}

	file := filepath.Join(strings.TrimSuffix(path, "\n"), binaryName)
	if ft == tarGz {
		err = downloadTarGz(url, file)
	} else {
		err = downloadBinary(url, file)
	}

	if err != nil {
		return msg(err)
	}

	fmt.Fprintf(os.Stdout, "Install '%s' to '%s'.\n", binaryName, file)
	return 0
}

func determinePath() string {
	if len(path) > 0 {
		return path
	}

	if len(cfg.Path) > 0 {
		return cfg.Path
	}

	if runtime.GOOS == "windows" {
		return "."
	}
	return "/usr/local/bin/"
}

func findDownloadURL(userName, repo string) (string, fileType, error) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), userName, repo)
	if err != nil {
		return "", unknown, err
	}

	logger.Printf("latest release version: %+v\n", *release.Name)

	for _, asset := range release.Assets {
		if isAvailableBinary(asset) {
			if strings.HasSuffix(*asset.Name, "tar.gz") {
				return *asset.BrowserDownloadURL, tarGz, nil
			}
			return *asset.BrowserDownloadURL, binary, nil
		}
	}

	msg := fmt.Sprintf("can't find an available released binary. isn't the binary name '%s'?", binaryName)
	return "", unknown, errors.New(msg)
}

func isAvailableBinary(asset github.ReleaseAsset) bool {
	if !isSupportedFormat(*asset.Name) {
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

func downloadTarGz(url, file string) error {
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	archive, err := gzip.NewReader(resp.Body)
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

func downloadBinary(url, file string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, bs, 0755)
	if err != nil {
		return err
	}
	return nil
}

func isSupportedFormat(name string) bool {
	// TODO(y-yagi): Support zip.
	suffixes := []string{"deb", "rpm", "msi", "zip"}
	for _, v := range suffixes {
		if strings.HasSuffix(name, v) {
			return false
		}
	}

	return true
}
