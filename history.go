package main

type history struct {
	URL  string
	Tag  string
	Path string
}

func (h *history) key() string {
	return h.Path
}
