package main

type history struct {
	URL        string
	Tag        string
	Path       string
	BinaryName string
}

func (h *history) key() string {
	return h.Path
}
