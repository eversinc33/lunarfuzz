package fuzz

type Result struct {
	Path    string
	Match   bool
	IsError bool
	Counter int
	Words   int
	Size    int
}
