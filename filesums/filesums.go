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
	Progress    bool
	Stats       *Stats
}

// Add searches for new files and adds them without also verifying existing files
func (self *FileSums) Add() (err error) {
	log.Info("Adding new file checksums")

	err = self.addOrUpdate(true)

	return
}

// Update searches for new files or files with wrong checksums and store them
func (self *FileSums) Update() (err error) {
	log.Info("Updating checksums")

	err = self.addOrUpdate(false)

	return
}

// Verify searches for previously seen files that do not match and report them
func (self *FileSums) Verify() (err error) {
	log.Info("Verifying checksums")

	wg := &sync.WaitGroup{}
	work := make(chan *Directory, len(self.Directories))
	failed := make(chan string, len(self.Directories))

	defer close(failed)

	for _, dir := range self.Directories {
		work <- dir
	}
	close(work)

	if self.Progress {
		go self.Stats.ShowProgress()
		defer self.Stats.StopProgress()
	}

	for w := 1; w <= self.Workers; w++ {
		wg.Add(1)
		go self.verifyWorker(work, failed, w, wg)
	}

	wg.Wait()

	return
}

func (self *FileSums) addOrUpdate(skipExisting bool) (err error) {
	wg := &sync.WaitGroup{}
	work := make(chan *Directory, len(self.Directories))

	for _, dir := range self.Directories {
		work <- dir
	}
	close(work)

	if self.Progress {
		go self.Stats.ShowProgress()
		defer self.Stats.StopProgress()
	}

	for w := 1; w <= self.Workers; w++ {
		wg.Add(1)
		if skipExisting {
			go self.addWorker(work, w, wg)
		} else {
			go self.updateWorker(work, w, wg)
		}
	}

	wg.Wait()

	return
}

// loads a checksum file or returns a empty checksum if none exist
func (self *FileSums) loadDirectory(directory string) (*Directory, error) {
	dir := NewDirectory(directory, !self.Progress)
	sumfile := path.Join(directory, ".checksums.json")

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
		j.Verify(failed, self.Stats)
	}
}

func (self *FileSums) updateWorker(jobs chan *Directory, worker int, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		err := j.Update(self.Stats)
		if err != nil {
			log.Errorf("Could not update %s: %s", j.path, err.Error())
		}
	}
}

func (self *FileSums) addWorker(jobs chan *Directory, worker int, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		err := j.Add(self.Stats)
		if err != nil {
			log.Errorf("Could not add %s: %s", j.path, err.Error())
		}
	}
}
