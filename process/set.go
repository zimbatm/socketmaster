package process

type Set map[*Process]bool

// A process set
func NewSet() Set {
	return Set(make(map[*Process]bool))
}

func (self Set) Add(p *Process) {
	if p != nil {
		self[p] = true
	}
}

func (self Set) Remove(p *Process) {
	delete(self, p)
}

func (self Set) KillAll() {
	for p, _ := range self {
		p.Kill()
	}
}

func (self Set) Empty() bool {
	return len(self) == 0
}

func (self Set) ToList() (l []*Process) {
	l = make([]*Process, len(self))
	i := 0
	for p, _ := range self {
		l[i] = p
		i++
	}
	return
}

func (self Set) Each(fn func(*Process)) {
	for p, _ := range self {
		fn(p)
	}
}
