package filesums

import (
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
)

type Stats struct {
	dircount     int
	directories  int
	new          int
	updated      int
	verified     int
	failed       int
	newFiles     []*string
	updatedFiles []*string
	failedFiles  []*string
	StartTime    time.Time
	bar          *uiprogress.Bar
	mu           *sync.Mutex
}

func NewStats() *Stats {
	return &Stats{
		StartTime: time.Now(),
		mu:        &sync.Mutex{},
	}
}

func (self *Stats) StopProgress() {
	self.bar.Set(self.Directories())
	uiprogress.Stop()
}

func (self *Stats) ShowProgress() {
	uiprogress.RefreshInterval = time.Millisecond * 1
	uiprogress.Start()

	self.bar = uiprogress.AddBar(self.DirCount())
	self.bar.AppendCompleted()

	self.bar.PrependFunc(func(b *uiprogress.Bar) string {
		return "Completed Sub Directories"
	})

	go func() {
		for {
			self.bar.Set(self.Directories())

			if self.IsCompleted() {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (self *Stats) FailedFiles() []*string {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.failedFiles
}

func (self *Stats) NewFiles() []*string {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.newFiles
}

func (self *Stats) UpdatedFiles() []*string {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.updatedFiles
}

func (self *Stats) IsCompleted() bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.directories >= self.dircount
}

func (self *Stats) DirCount() int {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.dircount
}

func (self *Stats) Directories() int {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.directories
}

func (self *Stats) New() int {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.new
}

func (self *Stats) Updated() int {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.updated
}

func (self *Stats) Verified() int {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.verified
}

func (self *Stats) Failed() int {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.failed
}

func (self *Stats) SetDirCount(c int) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.dircount = c
}

func (self *Stats) IncrDirectories() {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.directories++
}

func (self *Stats) IncrNew(f *string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.new++
	self.newFiles = append(self.newFiles, f)
}

func (self *Stats) IncrUpdated(f *string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.updated++
	self.updatedFiles = append(self.updatedFiles, f)
}

func (self *Stats) IncrVerified() {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.verified++
}

func (self *Stats) IncrFailed(f *string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.failed++
	self.failedFiles = append(self.failedFiles, f)
}
