package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"

	"github.com/y-yagi/goext/osext"
)

type HistoryFile struct {
	filename string
}

func (hf *HistoryFile) load() (map[string]*History, error) {
	if !osext.IsExist(hf.filename) {
		return nil, errors.New("history file doesn't exist")
	}

	b, err := os.ReadFile(hf.filename)
	if err != nil {
		return nil, err
	}

	var histories map[string]*History

	buf := bytes.NewBuffer(b)
	err = gob.NewDecoder(buf).Decode(&histories)
	if err != nil {
		return nil, err
	}

	return histories, nil
}

func (hf *HistoryFile) save(d Downloader, url, downloadedFile, binaryName string) error {
	var histories map[string]*History
	var buf *bytes.Buffer
	var err error

	if osext.IsExist(hf.filename) {
		histories, err = hf.load()
		if err != nil {
			return err
		}
	} else {
		histories = map[string]*History{}
	}

	h := History{URL: url, Tag: d.releaseTag, Path: downloadedFile, BinaryName: binaryName}
	histories[h.key()] = &h

	buf = bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(histories); err != nil {
		return err
	}

	return os.WriteFile(hf.filename, buf.Bytes(), 0600)
}
