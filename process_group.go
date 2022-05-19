package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"
)

type ProcessGroup struct {
	set *processSet
	wg  sync.WaitGroup

	inputs   Inputs
	sockfile *os.File
	user     *user.User
}

type processSet struct {
	sync.Mutex
	set map[*os.Process]bool
}

func MakeProcessGroup(inputs Inputs, sockfile *os.File, u *user.User) *ProcessGroup {
	return &ProcessGroup{
		set:      newProcessSet(),
		inputs:   inputs,
		sockfile: sockfile,
		user:     u,
	}
}

func (self *ProcessGroup) StartProcess() (process *os.Process, err error) {
	self.wg.Add(1)

	ioReader, ioWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	// Make sure parent values don't interfere with our child.
	// All fds have already been consumed at this stage.
	os.Unsetenv("LISTEN_PID")

	envMap := EnvironmentToMap(os.Environ())

	envMap["EINHORN_FDS"] = "3"
	envMap["LISTEN_FDS"] = "1"
	envMap["LISTEN_FDNAMES"] = "socket"

	env := MapToEnvironment(envMap)

	procAttr := &os.ProcAttr{
		Env:   env,
		Files: []*os.File{os.Stdin, ioWriter, ioWriter, self.sockfile},
		Sys:   &syscall.SysProcAttr{},
	}

	if self.user != nil {
		uid, _ := strconv.Atoi(self.user.Uid)
		gid, _ := strconv.Atoi(self.user.Gid)

		procAttr.Sys.Credential = &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		}
	}

	var commandPath string
	config, err := self.inputs.LoadConfig()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not load config '%s'", err))
	}

	commandPath, err = exec.LookPath(config.Command)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not find executable '%s'", err))
	}

	args := append([]string{LISTEN_PID_HELPER_ARGV0}, commandPath)
	args = append(args, flag.Args()...)
	log.Println("Starting", args[1:])
	process, err = os.StartProcess(os.Args[0], args, procAttr)
	if err != nil {
		return
	}

	// Add to set
	self.set.Add(process)

	// Prefix stdout and stderr lines with the [pid] and send it to the log
	logOutput(ioReader, process.Pid, &self.wg)

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

// A thread-safe process set
func newProcessSet() *processSet {
	set := new(processSet)
	set.set = make(map[*os.Process]bool)
	return set
}

func (self *processSet) Add(process *os.Process) {
	self.Lock()
	defer self.Unlock()

	self.set[process] = true
}

func (self *processSet) Each(fn func(*os.Process)) {
	self.Lock()
	defer self.Unlock()

	for process, _ := range self.set {
		fn(process)
	}
}

func (self *processSet) Remove(process *os.Process) {
	self.Lock()
	defer self.Unlock()
	delete(self.set, process)
}

func (self *processSet) Len() int {
	self.Lock()
	defer self.Unlock()
	return len(self.set)
}

func logOutput(input *os.File, pid int, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		var (
			err    error
			line   string
			reader = bufio.NewReader(input)
		)

		for err == nil {
			line, err = reader.ReadString('\n')
			if line != "" {
				log.Printf("[%d] %s", pid, line)
			}
		}
	}()
}
