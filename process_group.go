package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"syscall"
)

type ProcessGroup struct {
	set *ProcessSet
	wg  sync.WaitGroup
}

func NewProcessGroup() *ProcessGroup {
	return &ProcessGroup{
		set: NewProcessSet(),
	}
}

func (self *ProcessGroup) StartProcess(c *ProcessConfig) (process *os.Process, err error) {
	self.wg.Add(1)

	ioReader, ioWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	env := os.Environ()
	files := []*os.File{os.Stdin, ioWriter, ioWriter}
	command := c.command
	args := []string{c.command}

	if len(c.files) > 0 {
		// Einhorn compat
		env = append(env, fmt.Sprintf("EINHORN_MASTER_PID=%d", os.Getpid()))
		env = append(env, fmt.Sprintf("EINHORN_FD_COUNT=%d", len(c.files)))
		for i, _ := range c.files {
			env = append(env, fmt.Sprintf("EINHORN_FD_%d=%d", i, i+3))
		}
		// SystemD socket activation, LISTEN_PID below
		env = append(env, fmt.Sprintf("LISTEN_FDS=%d", len(c.files)))
		files = append(files, c.files...)
		command = "/bin/sh"
		args = []string{"/bin/sh", "-c", fmt.Sprintf("LISTEN_PID=$$ exec %s", c.command)}
	}

	procAttr := &os.ProcAttr{
		Env:   env,
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	}

	if c.user != nil {
		uid, _ := strconv.Atoi(c.user.Uid)
		gid, _ := strconv.Atoi(c.user.Gid)

		procAttr.Sys.Credential = &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		}
	}

	process, err = os.StartProcess(command, args, procAttr)
	if err != nil {
		return
	}

	// Add to set
	self.set.Add(process)

	// Prefix stdout and stderr lines with the [pid] and send it to the log
	go logOutput(ioReader, process.Pid, self.wg)

	// Handle the process death
	go func() {
		state, err := process.Wait()

		log.Println(process.Pid, state, err)

		// Remove from set
		self.set.Remove(process)

		// Process is gone
		ioReader.Close()
		self.wg.Done()
	}()

	return
}

func (self *ProcessGroup) SignalAll(signal os.Signal, except *os.Process) {
	self.set.Each(func(process *os.Process) {
		if process != except {
			process.Signal(signal)
		}
	})
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
