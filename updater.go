package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

type Updater struct {
	stdout          io.Writer
	stderr          io.Writer
	historyFilePath string
	cachePath       string
}

func (u *Updater) execute() error {
	hf := HistoryFile{filename: u.historyFilePath}
	histories, err := hf.load()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, history := range histories {
		go func(h *History) {
			defer wg.Done()

			parsedURL := strings.Split(h.URL, "/")
			downloader := Downloader{user: parsedURL[len(parsedURL)-2], repository: parsedURL[len(parsedURL)-1], binaryName: h.BinaryName, cachePath: u.cachePath, releaseTag: ""}

			err := downloader.findDownloadURL()
			if err != nil {
				mu.Lock()
				fmt.Fprintf(u.stderr, "An error occurred while updating '%v', '%v'\n", h.Path, err)
				mu.Unlock()
				return
			}

			if h.Tag == downloader.releaseTag {
				mu.Lock()
				fmt.Fprintf(u.stdout, "'%v' is already the latest version\n", h.Path)
				mu.Unlock()
				return
			}

			err = downloader.execute(h.Path)
			if err != nil {
				mu.Lock()
				fmt.Fprintf(u.stderr, "An error occurred while updating '%v', '%v'\n", h.Path, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			fmt.Fprintf(u.stdout, "Updated '%v' from '%v' to '%v'\n", h.Path, h.Tag, downloader.releaseTag)
			// TODO: Run save just once.
			hf.save(downloader, h.URL, h.Path, h.BinaryName)
			mu.Unlock()
		}(history)
		wg.Add(1)
	}

	wg.Wait()
	return nil
}
