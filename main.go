package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/y-yagi/configure"
	"github.com/y-yagi/debuglog"
	"github.com/y-yagi/goext/osext"
)

const cmd = "obt"

var (
	cfg    config
	logger *debuglog.Logger

	flags         *flag.FlagSet
	showVersion   bool
	showInstalled bool
	path          string
	defaultPath   string
	binaryName    string
	releaseTag    string
	historyFile   string

	version = "devel"
)

type config struct {
	Path      string `toml:"path"`
	CachePath string `toml:"cache_path"`
}

func main() {
	setFlags()
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func setFlags() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.BoolVar(&showVersion, "v", false, "print version number")
	flags.BoolVar(&showInstalled, "installed", false, "show installed binaries")
	flags.StringVar(&path, "p", "", "install path")
	flags.StringVar(&defaultPath, "s", "", "set default install path")
	flags.StringVar(&binaryName, "b", "", "binary name")
	flags.StringVar(&releaseTag, "tag", "", "release tag")
	flags.StringVar(&historyFile, "history", "", "history file")
	flags.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n\n", cmd)
	fmt.Fprintf(os.Stderr, "Install binary file from GitHub's release page. Default install path is '%s'.\n\n", determinePath())
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flags.PrintDefaults()
}

func msg(err error, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", cmd, err)
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) int {
	logger = debuglog.New(stdout)
	configure.Load(cmd, &cfg)

	flags.Parse(args[1:])

	if showVersion {
		fmt.Fprintf(stdout, "%s %s\n", cmd, version)
		return 0
	}

	if showInstalled {
		return msg(showInstalledBinaries(stdout), stderr)
	}

	if len(defaultPath) > 0 {
		cfg.Path = defaultPath
		configure.Save(cmd, cfg)
		fmt.Fprintf(stdout, "Change default install path to '%s'\n", defaultPath)
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
		return msg(err, stderr)
	}

	path := determinePath()
	if _, err := os.Stat(path); err != nil {
		return msg(err, stderr)
	}

	file := filepath.Join(strings.TrimSuffix(path, "\n"), downloader.binaryName)

	if osext.IsExist(file) {
		fmt.Fprintf(stdout, "'%s' exists. Override a file?\nPlease type (y)es or (n)o and then press enter: ", file)
		if !askForConfirmation(stdout) {
			fmt.Fprint(stdout, "download canceled.\n")
			return 0
		}
	}

	err = downloader.execute(file)
	if err != nil {
		return msg(err, stderr)
	}

	err = saveHistory(&downloader, file, url)
	if err != nil {
		fmt.Fprintf(stderr, "history save error %v\n", err)
	}
	fmt.Fprintf(stdout, "Download '%s(%s)' to '%s'.\n", downloader.binaryName, downloader.releaseTag, file)
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

func askForConfirmation(stdout io.Writer) bool {
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
		fmt.Fprintln(stdout, "Please type (y)es or (n)o and then press enter: ")
		return askForConfirmation(stdout)
	}
}

func saveHistory(d *downloader, fullpath, url string) error {
	var histories map[string]*history
	var buf *bytes.Buffer

	if len(historyFile) == 0 {
		historyFile = filepath.Join(configure.ConfigDir(cmd), "history")
	}

	if osext.IsExist(historyFile) {
		b, err := ioutil.ReadFile(historyFile)
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

	h := history{URL: url, Tag: d.releaseTag, Path: fullpath}
	histories[h.key()] = &h

	buf = bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(histories)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(historyFile, buf.Bytes(), 0644)
}

func showInstalledBinaries(stdout io.Writer) error {
	var histories map[string]*history

	if len(historyFile) == 0 {
		historyFile = filepath.Join(configure.ConfigDir(cmd), "history")
	}

	if !osext.IsExist(historyFile) {
		return errors.New("history file doesn't exist")
	}

	b, err := ioutil.ReadFile(historyFile)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	err = gob.NewDecoder(buf).Decode(&histories)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(stdout)
	table.SetHeader([]string{"URL", "TAG", "PATH"})

	for _, h := range histories {
		table.Append([]string{h.URL, h.Tag, h.Path})
	}

	table.Render()
	return nil
}
