package main

type History struct {
	URL        string
	Tag        string
	Path       string
	BinaryName string
}

func (h *History) key() string {
	return h.Path
}
