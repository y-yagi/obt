# obt

`obt` is a downloader for the GitHub release page. This tool automatically downloads the latest release file that according to OS and architecture.

## Example

```bash
$ obt https://github.com/davecheney/httpstat
Install 'httpstat' to '/home/y-yagi/.local/bin/httpstat'.
$ obt -p /usr/local/bin https://github.com/rclone/rclone
Install 'rclone' to '/usr/local/bin/rclone'.
$ obt -b staticcheck https://github.com/dominikh/go-tools
Install 'staticcheck' to '/home/y-yagi/.local/bin/staticcheck'.
```

## Usage

```
$ obt --help
Usage: obt [OPTIONS] URL

Install binary file from GitHub's release page. Default install path is '/usr/local/bin/'.

OPTIONS:
  -b string
    	binary name(default: repository name)
  -p string
    	install path
  -s string
    	set default install path
  -v	print version number
```

## Default install path

`obt` uses `/usr/local/bin/` to a default install path in case of Linux or macOS. In windows, uses `.`.

You can change a default install path via `-s` option.

```bash
obt -s /home/y-yagi/.local/bin/
Change default install path to '/home/y-yagi/.local/bin/'
```

## Installation

Download files from [GitHub release page](https://github.com/y-yagi/obt/releases).
