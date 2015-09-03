package process

import (
	"os"
	"sync"
)

type Set struct {
	sync.Mutex
	set map[*os.Process]bool
}

// A thread-safe process set
func NewSet() *Set {
	set := new(Set)
	set.set = make(map[*os.Process]bool)
	return set
}

func (self *Set) Add(p *os.Process) {
	self.Lock()
	defer self.Unlock()
	self.set[p] = true
}

func (self *Set) Each(fn func(*os.Process)) {
	self.Lock()
	defer self.Unlock()

	for p, _ := range self.set {
		fn(p)
	}
}

func (self *Set) Remove(p *os.Process) {
	self.Lock()
	defer self.Unlock()
	delete(self.set, p)
}
