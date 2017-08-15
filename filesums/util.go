package filesums

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func isDir(path string) bool {
	dirInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	if !dirInfo.IsDir() {
		return false
	}

	return true
}

func hasFile(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	return true
}

func computeMd5(filePath string) (string, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(result)), nil
}

func loadChecksums(path string, dir *Directory) error {
	if !hasFile(path) {
		return fmt.Errorf("Could not load checksums from %s: file does not exist", path)
	}

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Could not open checksum file %s: %s", path, err.Error())
	}

	if err := json.Unmarshal(dat, dir); err != nil {
		return fmt.Errorf("Could not parse JSON data in file %s: %s", path, err.Error())
	}

	return nil
}
