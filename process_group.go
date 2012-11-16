package main

import (
	"bufio"
	"fmt"
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

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	env := append(os.Environ(), "EINHORN_FDS=3")
	procattr := &os.ProcAttr{
		Env:   env,
		Files: []*os.File{os.Stdin, stdoutWriter, stderrWriter, self.sockfile},
	}

	process, err = os.StartProcess(self.commandPath, []string{}, procattr)
	if err != nil {
		return
	}

	// Add to set
	self.set[process] = true

	// Helps waiting for process, stdout and stderr
	var wg sync.WaitGroup

	// Prefix stdout and stderr lines with the [pid]
	go PrefixOutput(stdoutReader, os.Stdout, process.Pid, wg)
	go PrefixOutput(stderrReader, os.Stderr, process.Pid, wg)

	// Handle the process death
	go func() {
		wg.Add(1)

		state, err := process.Wait()

		log.Println(process.Pid, state, err)

		// Remove from set
		delete(self.set, process)

		// Process is gone
		stdoutReader.Close()
		stderrReader.Close()
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

func PrefixOutput(input *os.File, output *os.File, pid int, wg sync.WaitGroup) {
	var err error
	var line string
	wg.Add(1)

	reader := bufio.NewReader(input)

	for err == nil {
		line, err = reader.ReadString('\n')
		if line != "" {
			output.WriteString(fmt.Sprintf("[%d] %s", pid, line))
		}
	}

	wg.Done()
}
