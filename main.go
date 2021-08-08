package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/y-yagi/configure"
	"github.com/y-yagi/debuglog"
	"github.com/y-yagi/goext/osext"
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
	releaseTag  string

	version = "devel"
)

type config struct {
	Path      string `toml:"path"`
	CachePath string `toml:"cache_path"`
}

type history struct {
	URL  string
	Tag  string
	Path string
}

func main() {
	setFlags()
	os.Exit(run(os.Args))
}

func setFlags() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.BoolVar(&showVersion, "v", false, "print version number")
	flags.StringVar(&path, "p", "", "install path")
	flags.StringVar(&defaultPath, "s", "", "set default install path")
	flags.StringVar(&binaryName, "b", "", "binary name")
	flags.StringVar(&releaseTag, "tag", "", "release tag")
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
		fmt.Fprintf(os.Stdout, "%s %s\n", cmd, version)
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

	if len(a) < 2 {
		flags.Usage()
		return 0
	}

	if len(cfg.CachePath) == 0 {
		dir, err := os.UserCacheDir()
		if err == nil {
			cfg.CachePath = filepath.Join(dir, cmd)
		}
	}

	downloader := downloader{user: a[len(a)-2], repository: a[len(a)-1], binaryName: binaryName, cachePath: cfg.CachePath, releaseTag: releaseTag}
	err := downloader.findDownloadURL()
	if err != nil {
		return msg(err)
	}

	path := determinePath()
	if _, err := os.Stat(path); err != nil {
		return msg(err)
	}

	file := filepath.Join(strings.TrimSuffix(path, "\n"), downloader.binaryName)

	if osext.IsExist(file) {
		fmt.Fprintf(os.Stdout, "'%s' exists. Override a file?\nPlease type (y)es or (n)o and then press enter: ", file)
		if !askForConfirmation() {
			fmt.Fprint(os.Stdout, "download canceled.\n")
			return 0
		}
	}

	err = downloader.execute(file)
	if err != nil {
		return msg(err)
	}

	err = saveHistory(&downloader, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "history save error %v\n", err)
	}
	fmt.Fprintf(os.Stdout, "Download '%s(%s)' to '%s'.\n", downloader.binaryName, downloader.releaseTag, file)
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

func askForConfirmation() bool {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		fmt.Fprintln(os.Stdout, "Please type (y)es or (n)o and then press enter: ")
		return askForConfirmation()
	}
}

func saveHistory(d *downloader, url string) error {
	var histories map[string]*history
	var buf *bytes.Buffer
	filename := filepath.Join(configure.ConfigDir(cmd), "history")

	if osext.IsExist(filename) {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(b)
		err = gob.NewDecoder(buf).Decode(&histories)
		if err != nil {
			return err
		}
	} else {
		histories = map[string]*history{}
	}

	h := history{URL: url, Tag: d.releaseTag, Path: d.binaryName}
	histories[h.key()] = &h

	buf = bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(histories)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

func (h *history) key() string {
	return h.Path
}
