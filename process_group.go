package main

import (
	"bufio"
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

	ioReader, ioWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	env := append(os.Environ(), "EINHORN_FDS=3")
	procattr := &os.ProcAttr{
		Env:   env,
		Files: []*os.File{os.Stdin, ioWriter, ioWriter, self.sockfile},
	}

	process, err = os.StartProcess(self.commandPath, []string{}, procattr)
	if err != nil {
		return
	}

	// Add to set
	self.set[process] = true

	// Helps waiting for process, stdout and stderr
	var wg sync.WaitGroup

	// Prefix stdout and stderr lines with the [pid] and send it to the log
	go logOutput(ioReader, process.Pid, wg)

	// Handle the process death
	go func() {
		wg.Add(1)

		state, err := process.Wait()

		log.Println(process.Pid, state, err)

		// Remove from set
		delete(self.set, process)

		// Process is gone
		ioReader.Close()
		wg.Done()
	}()

	// Wait for process, stdout and stderr before declaring the process done
	go func() {
		wg.Wait()
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

func logOutput(input *os.File, pid int, wg sync.WaitGroup) {
	var err error
	var line string
	wg.Add(1)

	reader := bufio.NewReader(input)

	for err == nil {
		line, err = reader.ReadString('\n')
		if line != "" {
			log.Printf("[%d] %s", pid, line)
		}
	}

	wg.Done()
}
