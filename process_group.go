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

	// For tracking which processes run an up to date config.
	// Incremented on SIGHUP.
	generation int

	inputs   Inputs
	sockfile *os.File
	user     *user.User
}

// Below is the process life cycle state machine.
//
// It is not just descriptive, but prescriptive, as unexpected interleavings
// of events must not put socketmaster in an unexpected state.
//
// From \ To   | Starting     Operational  Yielding     Yielded      Stopping      Gone
// ------------+----------------------------------------------------------------------------------
// Starting    | id           happy                                  early stop    startup-fail
// Operational |              id           happy                     shutdown      operational-fail
// Yielding    |                           id           happy        yield-timeout operational-fail
// Yielded     |                                        id           last-call     exit a or operational-fail
// Stopping    |                                                     id            exit b or killed or operational-fail
// Gone        |                                                                   id
//
// id:               identity aka no-op
// happy:            the usual path
// early stop:       startup took too long and/or socketmaster wasn't notified that the process was ready
// shutdown:         socketmaster is shutting down, not doing a hot reload
// yield-timeout:    if the process doesn't respond to the request to yield, it does not know how to yield, or is defunct, so we stop and kill
// last-call:        at some point we can't allow the old process to remain, so we stop and kill
// exit a:           voluntary exit after being asked to yield
// exit b:           graceful exit after being asked to yield and asked to stop
// *-fail:           what it says on the tin
//
// Note that the lower triangle is empty, so the state machine has no cycles
// besides id. "Cyclical" behavior only manifests at a higher level, as new
// processes replace old ones.
type ProcessLifecycleState int

const (
	Starting    ProcessLifecycleState = iota
	Operational                       // Accepting connections and handling existing connections
	Yielding                          // Operational, should stop accepting connections
	Yielded                           // Only handling existing connections
	Stopping                          // releasing resources, cleaning up
	Gone
)

type ProcessState struct {
	sync.Mutex
	generation     int
	lifecycleState ProcessLifecycleState
}

func (self *ProcessState) CanStop() bool {
	return self.lifecycleState == Operational || self.lifecycleState == Stopping
}

type processSet struct {
	sync.Mutex
	set map[*os.Process]ProcessState
}

func MakeProcessGroup(inputs Inputs, sockfile *os.File, u *user.User) *ProcessGroup {
	return &ProcessGroup{
		set:        newProcessSet(),
		inputs:     inputs,
		sockfile:   sockfile,
		user:       u,
		generation: 0,
	}
}

func (self *ProcessGroup) StartProcess() (process *os.Process, err error) {
	self.wg.Add(1)

	ioReader, ioWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	config, err := self.inputs.LoadConfig()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not load config '%s'", err))
	}

	// Make sure parent values don't interfere with our child.
	// All fds have already been consumed at this stage.
	os.Unsetenv("LISTEN_PID")

	envMap := EnvironmentToMap(os.Environ())

	envMap["EINHORN_FDS"] = "3"
	envMap["LISTEN_FDS"] = "1"
	envMap["LISTEN_FDNAMES"] = "socket"

	for key, value := range config.Environment {
		envMap[key] = value
	}

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

	state := ProcessState{
		generation:     self.generation,
		lifecycleState: Starting,
	}

	// Add to set
	self.set.Add(process, state)

	// Prefix stdout and stderr lines with the [pid] and send it to the log
	logOutput(ioReader, process.Pid, &self.wg)

	// Handle the process death
	go func() {
		osProcState, err := process.Wait()

		log.Println(process.Pid, osProcState, err)

		// Remove from set
		self.set.Lock()
		self.set.Remove(process)
		self.set.Unlock()

		// Process is gone
		ioReader.Close()
		self.wg.Done()

		state.Lock()
		state.lifecycleState = Gone
		state.Unlock()
	}()

	return
}

func (self *ProcessGroup) SignalAll(signal os.Signal, maxGeneration int) {
	self.set.Each(func(process *os.Process, state ProcessState) {
		if state.generation <= maxGeneration {
			process.Signal(signal)
		}
	})
}

func (self *ProcessGroup) TerminateAll(signal os.Signal, maxGeneration int) {
	self.set.Each(func(process *os.Process, state ProcessState) {
		state.Lock()
		if state.generation <= maxGeneration && state.CanStop() {
			process.Signal(signal)
			state.lifecycleState = Stopping
		}
		state.Unlock()
	})
}

func (self *ProcessGroup) WaitAll() {
	self.wg.Wait()
}

// A thread-safe process set
func newProcessSet() *processSet {
	set := new(processSet)
	set.set = make(map[*os.Process]ProcessState)
	return set
}

func (self *processSet) Add(process *os.Process, state ProcessState) {
	self.Lock()
	defer self.Unlock()

	self.set[process] = state
}

func (self *processSet) Each(fn func(*os.Process, ProcessState)) {
	self.Lock()
	defer self.Unlock()

	for process, state := range self.set {
		fn(process, state)
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
