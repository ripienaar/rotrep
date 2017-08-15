package filesums

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type FileSums struct {
	Directories map[string]*Directory
	Root        string
	Workers     int
	Quiet       bool
}

// Update searches for new files or files with wrong checksums and store them
func (self *FileSums) Update() (changed int, added int, err error) {
	log.Info("Updating checksums")

	work := make(chan *Directory, len(self.Directories))
	updated := make(chan string, len(self.Directories))
	new := make(chan string, len(self.Directories))

	wg := &sync.WaitGroup{}

	for _, dir := range self.Directories {
		work <- dir
	}
	close(work)

	for w := 1; w <= self.Workers; w++ {
		wg.Add(1)
		go self.updateWorker(work, updated, new, w, wg)
	}

	go func() {
		for f := range new {
			if !self.Quiet {
				fmt.Printf("new: %s\n", f)
			}
			added++
		}
	}()

	go func() {
		for f := range updated {
			if !self.Quiet {
				fmt.Printf("updated: %s\n", f)
			}
			changed++
		}
	}()

	wg.Wait()

	close(updated)
	close(new)

	return
}

// Verify searches for previously seen files that do not match and report them
func (self *FileSums) Verify() (fcount int, err error) {
	log.Info("Verifying checksums")

	work := make(chan *Directory, len(self.Directories))
	failed := make(chan string, len(self.Directories))
	wg := &sync.WaitGroup{}

	for _, dir := range self.Directories {
		work <- dir
	}
	close(work)

	for w := 1; w <= self.Workers; w++ {
		wg.Add(1)
		go self.verifyWorker(work, failed, w, wg)
	}

	go func() {
		for f := range failed {
			if !self.Quiet {
				fmt.Printf("failed: %s\n", f)
			}

			fcount++
		}
	}()

	wg.Wait()

	close(failed)

	return
}

// loads a checksum file or returns a empty checksum if none exist
func (self *FileSums) loadDirectory(directory string) (*Directory, error) {
	dir := NewDirectory(directory)
	sumfile := path.Join(directory, "checksums.json")

	if !isDir(directory) {
		return dir, fmt.Errorf("Could not load '%s', it's not a directory", directory)
	}

	if !hasFile(sumfile) {
		return dir, nil
	}

	if err := loadChecksums(sumfile, dir); err != nil {
		return dir, err
	}

	return dir, nil
}

func (self *FileSums) shouldCheckDir(file string) bool {
	_, dir := path.Split(file)

	if strings.HasPrefix(dir, ".") {
		return false
	}

	return true
}

// walks the tree and record all directory names
func (self *FileSums) populateDirectories() error {
	log.Info("Populating subdirectories")

	filepath.Walk(self.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Could not traverse %s: %s", path, err.Error())
			return nil
		}

		if !self.shouldCheckDir(path) {
			return nil
		}

		if isDir(path) {
			dir, err := self.loadDirectory(path)
			if err != nil {
				return err
			}

			self.Directories[path] = dir
		}

		return nil
	})

	log.Infof("Found %d directories", len(self.Directories))

	return nil
}

func (self *FileSums) verifyWorker(jobs chan *Directory, failed chan string, worker int, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		j.Verify(failed)
	}
}

func (self *FileSums) updateWorker(jobs chan *Directory, updated chan string, new chan string, worker int, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		err := j.Update(updated, new)
		if err != nil {
			log.Errorf("Could not update %s: %s", j.path, err.Error())
		}
	}
}
