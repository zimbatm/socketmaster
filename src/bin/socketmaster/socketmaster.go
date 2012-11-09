package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"socketmaster"
	"sync"
	"syscall"
	"time"
)

/********* Process handling ********/

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

	procattr := &os.ProcAttr{
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

		fmt.Println(state, err)

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

func handleSignals(processGroup *ProcessGroup, c chan os.Signal, startTime int) {
	for {
		signal := <-c // os.Signal
		syscallSignal := signal.(syscall.Signal)

		switch syscallSignal {
		case syscall.SIGHUP:
			go func() {
				process, err := processGroup.StartProcess()
				if err != nil {
					fmt.Errorf("Could not start new process: %v", err)
				} else {
					time.Sleep(time.Duration(startTime) * time.Millisecond)

					// A possible improvement woud be to only swap the
					// process if the new child is still alive.
					processGroup.SignalAll(signal, process)
				}
			}()
		default:
			// Forward signal
			processGroup.SignalAll(signal, nil)
		}
	}
}

func main() {
	var addr string
	var err error
	var startTime int
	var command string

	flag.StringVar(&addr, "listen", "tcp://:8080", "Port on which to bind")
	flag.IntVar(&startTime, "start", 3000, "How long the new process takes to boot in millis")
	flag.StringVar(&command, "command", "", "Program to start")
	flag.Parse()

	if command == "" {
		log.Fatalln("Command path is mandatory")
	}

	commandPath, err := exec.LookPath(command)
	if err != nil {
		log.Fatalln("Could not find executable", err)
	}

	sockfile, err := socketmaster.ListenFile(addr)
	if err != nil {
		log.Fatalln("Unable to open socket", err)
	}

	// Run the first process
	processGroup := MakeProcessGroup(commandPath, sockfile)
	_, err = processGroup.StartProcess()
	if err != nil {
		log.Fatalln("Could not start process", err)
	}

	// Monitoring the processes
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go handleSignals(processGroup, c, startTime)

	// TODO: Full restart on USR2. Make sure the listener file is not set to SO_CLOEXEC
	// TODO: Restart processes if they die

	// For now, exit if no processes are left
	processGroup.WaitAll()
}
