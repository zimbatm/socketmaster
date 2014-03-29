package main

import (
	"os"
	"sync"
)

type ProcessSet struct {
	sync.Mutex
	set map[*os.Process]bool
}

// A thread-safe process set
func NewProcessSet() *ProcessSet {
	set := new(ProcessSet)
	set.set = make(map[*os.Process]bool)
	return set
}

func (self *ProcessSet) Add(process *os.Process) {
	self.Lock()
	defer self.Unlock()
	self.set[process] = true
}

func (self *ProcessSet) Each(fn func(*os.Process)) {
	self.Lock()
	defer self.Unlock()

	for process, _ := range self.set {
		fn(process)
	}
}

func (self *ProcessSet) Remove(process *os.Process) {
	self.Lock()
	defer self.Unlock()
	delete(self.set, process)
}
