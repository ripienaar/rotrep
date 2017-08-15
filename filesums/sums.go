package filesums

import (
	"fmt"
	"path/filepath"
)

func NewFileSums(root string, workers int, quiet bool, progress bool) (*FileSums, error) {
	sums := &FileSums{
		Root:        root,
		Directories: make(map[string]*Directory),
		Workers:     workers,
		Quiet:       quiet,
		Progress:    progress,
		Stats:       NewStats(),
	}

	a, err := filepath.Abs(root)
	if err != nil {
		return sums, fmt.Errorf("Could not determine absolute path of %s: %s", root, err.Error())
	}
	sums.Root = a

	if !isDir(a) {
		return sums, fmt.Errorf("Directory %s does not exist, cannot manage checksums", a)
	}

	if err := sums.populateDirectories(); err != nil {
		return sums, fmt.Errorf("Could not populate directories under %s: %s", root, err.Error())
	}

	sums.Stats.SetDirCount(len(sums.Directories))

	return sums, nil
}
