package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"

	"github.com/y-yagi/goext/osext"
)

type HistoryFile struct {
	filename string
}

func (hf *HistoryFile) load() (map[string]*history, error) {
	if !osext.IsExist(hf.filename) {
		return nil, errors.New("history file doesn't exist")
	}

	b, err := ioutil.ReadFile(hf.filename)
	if err != nil {
		return nil, err
	}

	var histories map[string]*history

	buf := bytes.NewBuffer(b)
	err = gob.NewDecoder(buf).Decode(&histories)
	if err != nil {
		return nil, err
	}

	return histories, nil
}

func (hf *HistoryFile) save(d downloader, url, downloadedFile, binaryName string) error {
	var histories map[string]*history
	var buf *bytes.Buffer
	var err error

	if osext.IsExist(hf.filename) {
		histories, err = hf.load()
		if err != nil {
			return err
		}
	} else {
		histories = map[string]*history{}
	}

	h := history{URL: url, Tag: d.releaseTag, Path: downloadedFile, BinaryName: binaryName}
	histories[h.key()] = &h

	buf = bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(histories); err != nil {
		return err
	}

	return ioutil.WriteFile(hf.filename, buf.Bytes(), 0644)
}
