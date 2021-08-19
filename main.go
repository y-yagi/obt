package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

	flags           *flag.FlagSet
	showVersion     bool
	showInstalled   bool
	path            string
	defaultPath     string
	binaryName      string
	releaseTag      string
	historyFilePath string

	version = "devel"
)

type config struct {
	Path            string `toml:"path"`
	CachePath       string `toml:"cache_path"`
	HistoryFilePath string `toml:"history_file_path"`
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
	flags.StringVar(&historyFilePath, "history", "", "set history file path")
	flags.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n\n", cmd)
	fmt.Fprintf(os.Stderr, "Install binary file from GitHub's release page. Default install path is '%s'.\n\n", cfg.Path)
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
		if !filepath.IsAbs(defaultPath) {
			fmt.Fprintln(stderr, "Please specify an absolute path to the default install path.")
			return 1
		}

		cfg.Path = defaultPath
		configure.Save(cmd, cfg)
		fmt.Fprintf(stdout, "Change default install path to '%s'\n", defaultPath)
		return 0
	}

	if len(historyFilePath) > 0 {
		if !filepath.IsAbs(historyFilePath) {
			fmt.Fprintln(stderr, "Please specify an absolute path to the history file path.")
			return 1
		}

		cfg.HistoryFilePath = historyFilePath
		configure.Save(cmd, cfg)
		fmt.Fprintf(stdout, "Change history file path to '%s'\n", historyFilePath)
		return 0
	}

	return msg(download(stdout, stderr), stderr)
}

func download(stdout, stderr io.Writer) error {
	if len(flags.Args()) == 0 {
		flags.Usage()
		return nil
	}

	url := strings.TrimSuffix(flags.Args()[0], "/")
	a := strings.Split(url, "/")

	if len(a) < 2 {
		flags.Usage()
		return nil
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
		return err
	}

	path, err := determinePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err != nil {
		return err
	}

	file := filepath.Join(strings.TrimSuffix(path, "\n"), downloader.binaryName)

	if osext.IsExist(file) {
		fmt.Fprintf(stdout, "'%s' exists. Override a file?\nPlease type (y)es or (n)o and then press enter: ", file)
		if !askForConfirmation(stdout) {
			fmt.Fprint(stdout, "download canceled.\n")
			return nil
		}
	}

	err = downloader.execute(file)
	if err != nil {
		return err
	}

	if len(path) == 0 {
		hf := HistoryFile{filename: determineHistoryFilePath()}
		err = hf.save(downloader, url, file)
		if err != nil {
			fmt.Fprintf(stderr, "history save error %v\n", err)
		}
	}

	fmt.Fprintf(stdout, "Download '%s(%s)' to '%s'.\n", downloader.binaryName, downloader.releaseTag, file)
	return nil
}

func determinePath() (string, error) {
	if len(path) > 0 {
		return path, nil
	}

	if len(cfg.Path) > 0 {
		return cfg.Path, nil
	}

	return "", errors.New("please set a default install path(via '-s option') or an install path(via '-p' option)")
}

func determineHistoryFilePath() string {
	if len(cfg.HistoryFilePath) > 0 {
		return filepath.Join(cfg.HistoryFilePath, "history")
	}

	return filepath.Join(configure.ConfigDir(cmd), "history")
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

func showInstalledBinaries(stdout io.Writer) error {
	hf := HistoryFile{filename: determineHistoryFilePath()}
	histories, err := hf.load()
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
