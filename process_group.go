package main

import (
	"log"
	"os"
	"sync"
)

type ProcessGroup struct {
	set map[*os.Process]bool
	wg  sync.WaitGroup

	commandPath string
	sockfile    *os.File
}

func MakeProcessGroup(commandPath string, sockfile *os.File) *ProcessGroup {
	pg := &ProcessGroup{
		set: make(map[*os.Process]bool),

		commandPath: commandPath,
		sockfile:    sockfile,
	}

	return pg
}

func (self *ProcessGroup) StartProcess() (process *os.Process, err error) {
	self.wg.Add(1)

	env := append(os.Environ(), "EINHORN_FDS=3")

	procattr := &os.ProcAttr{
		Env: env,
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr, self.sockfile},
	}

	process, err = os.StartProcess(self.commandPath, []string{}, procattr)
	if err != nil {
		return
	}

	// Add to set
	self.set[process] = true

	go func() {
		state, err := process.Wait()

		log.Println(state, err)

		// Remove from set
		delete(self.set, process)

		//
		self.wg.Done()
	}()

	return
}

func (self *ProcessGroup) SignalAll(signal os.Signal, except *os.Process) {
	for process, _ := range self.set {
		if process != except {
			process.Signal(signal)
		}
	}
}

func (self *ProcessGroup) WaitAll() {
	self.wg.Wait()
}
