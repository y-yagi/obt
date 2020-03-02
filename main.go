package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

	downloader := downloader{user: a[len(a)-2], repository: a[len(a)-1]}
	err := downloader.findDownloadURL()
	if err != nil {
		return msg(err)
	}

	logger.Printf("download file from '%+v'\n", downloader.url)
	path := determinePath()
	if _, err := os.Stat(path); err != nil {
		return msg(err)
	}

	file := filepath.Join(strings.TrimSuffix(path, "\n"), binaryName)
	err = downloader.execute(file)
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
