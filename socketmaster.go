package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func handleSignals(processGroup *ProcessGroup, c chan os.Signal, startTime int) {
	for {
		signal := <-c // os.Signal
		syscallSignal := signal.(syscall.Signal)

		switch syscallSignal {
		case syscall.SIGHUP:
			go func() {
				process, err := processGroup.StartProcess()
				if err != nil {
					log.Println("Could not start new process: %v", err)
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

	sockfile, err := ListenFile(addr)
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
