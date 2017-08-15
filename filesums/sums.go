package filesums

import (
	"fmt"
)

func NewFileSums(root string, workers int, quiet bool) (*FileSums, error) {
	sums := &FileSums{
		Root:        root,
		Directories: make(map[string]*Directory),
		Workers:     workers,
		Quiet:       true,
	}

	if !isDir(root) {
		return sums, fmt.Errorf("Directory %s does not exist, cannot manage checksums", root)
	}

	if err := sums.populateDirectories(); err != nil {
		return sums, fmt.Errorf("Could not populate directories under %s: %s", root, err.Error())
	}

	return sums, nil
}
