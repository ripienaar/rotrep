package filesums

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Directory struct {
	Created int64             `json:"created"`
	Updated int64             `json:"updated"`
	Files   map[string]string `json:"files"`
	path    string
}

func NewDirectory(path string) *Directory {
	return &Directory{
		Created: time.Now().Unix(),
		Updated: time.Now().Unix(),
		Files:   make(map[string]string),
		path:    path,
	}
}

func (self *Directory) save() error {
	out := filepath.Join(self.path, ".checksums.json")
	log.Debugf("Saving %s", out)

	self.Updated = time.Now().Unix()

	j, err := json.Marshal(self)
	if err != nil {
		return fmt.Errorf("Could not JSON encode %s: %s", self.path, err.Error())
	}

	err = ioutil.WriteFile(out, j, 0644)
	if err != nil {
		return fmt.Errorf("writing to %s failed: %s", out, err.Error())
	}

	return nil
}

func (self *Directory) shouldCheckFile(file string) bool {
	if strings.HasPrefix(file, ".checksums.json") {
		return false
	}

	return true
}

func (self *Directory) Verify(failed chan string) bool {
	success := true

	files, err := ioutil.ReadDir(self.path)
	if err != nil {
		err := fmt.Errorf("Could not update %s: %s", self.path, err.Error())
		fmt.Println(err.Error())
		panic(err)
	}

	for _, file := range files {
		fqpath := path.Join(self.path, file.Name())

		if isDir(fqpath) || !self.shouldCheckFile(file.Name()) {
			continue
		}

		log.Debugf("verifying file %s", fqpath)

		rsum, seen := self.Files[file.Name()]

		if !seen {
			log.Debugf("Skipping previously unseen file %s", fqpath)
			continue
		}

		csum, err := computeMd5(fqpath)
		if err != nil {
			err := fmt.Errorf("Could not calculate checksum for %s: %s", file.Name(), err.Error())
			fmt.Println(err.Error())
			panic(err)
		}

		if csum != rsum {
			failed <- fqpath
			success = false
			log.Warnf("Mismatch %s %s != %s", fqpath, csum, rsum)
		}
	}

	return success
}

func (self *Directory) Update(updated chan string, new chan string) error {
	found := false

	files, err := ioutil.ReadDir(self.path)
	if err != nil {
		err := fmt.Errorf("Could not update %s: %s", self.path, err.Error())
		fmt.Println(err.Error())
		panic(err)
	}

	for _, file := range files {
		fqpath := path.Join(self.path, file.Name())

		if isDir(fqpath) || !self.shouldCheckFile(file.Name()) {
			continue
		}

		log.Debugf("updating file %s", fqpath)

		csum, err := computeMd5(fqpath)
		if err != nil {
			err := fmt.Errorf("Could not calculate checksum for %s: %s", file.Name(), err.Error())
			fmt.Println(err.Error())
			panic(err)
		}

		rsum, seen := self.Files[file.Name()]

		if !seen {
			self.Files[file.Name()] = csum
			new <- fqpath
			found = true
			log.Debugf("Captured %s", fqpath)
		} else if csum != rsum {
			self.Files[file.Name()] = csum
			updated <- fqpath
			found = true
			log.Debugf("Updated %s %s -> %s", fqpath, rsum, csum)
		}
	}

	if found {
		err := self.save()
		if err != nil {
			return fmt.Errorf("Could not save checksums: %s", err.Error())
		}
	}

	return nil
}
