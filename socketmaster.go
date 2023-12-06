package main

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"
)

const PROGRAM_NAME = "socketmaster"

func handleSignals(processGroup *ProcessGroup, c <-chan os.Signal, startTime int, sockfile *os.File) {
	for {
		signal := <-c // os.Signal
		syscallSignal := signal.(syscall.Signal)

		switch syscallSignal {
		case syscall.SIGUSR1:
			socketMasterFdEnvVar := fmt.Sprintf("SOCKETMASTER_FD=%d", sockfile.Fd())
			syscall.Exec(os.Args[0], os.Args, append(os.Environ(), socketMasterFdEnvVar))
		case syscall.SIGHUP:
			oldGeneration := processGroup.generation
			process, err := processGroup.StartProcess()
			if err != nil {
				log.Printf("Could not start new process: %v\n", err)
			} else {
				if startTime > 0 {
					time.Sleep(time.Duration(startTime) * time.Millisecond)
				}

				if processGroup.set.Len() > 1 {
					processGroup.SignalAll(syscall.SIGTERM, oldGeneration)
				} else {
					log.Println("Failed to kill old process, because there's no one left in the group")
				}
			}
		default:
			// Forward signal
			processGroup.SignalAll(signal, nil)
		}
	}
}

// Go won't let us set LISTEN_PID between fork,
// and exec because letting "language users" do
// that is not necessarily safe, because of rts
// issues. Very understandable, but a little   annoying.
//
// So, instead we call ourselves with a special
// argv[0] that drops us into this little helper.
// That way we don't have to squeeze it between
// those syscalls, at the cost of an extra exec.
// These aren't hot execs, so performance is not
// really affected.
func setLISTEN_PIDHelper() {
	os.Setenv("LISTEN_PID", fmt.Sprint(os.Getpid()))
	args := os.Args[1:]
	syscall.Exec(args[0], args, os.Environ())
}

// See setLISTEN_PIDHelper
var LISTEN_PID_HELPER_ARGV0 = "set-LISTEN_PID-helper"

func main() {
	if os.Args[0] == LISTEN_PID_HELPER_ARGV0 {
		setLISTEN_PIDHelper()
	}

	inputs, err := ParseInputs(os.Args[1:])
	if err != nil {
		log.Fatalf("Options not valid: %s\n", err)
	}

	var config *Config
	config, err = inputs.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config file: %s\n", err)
	}

	useSyslog := inputs.useSyslog
	command := config.Command
	addr := inputs.addr
	username := inputs.username
	startTime := inputs.startTime

	if useSyslog {
		stream, err := syslog.New(syslog.LOG_INFO, PROGRAM_NAME)
		if err != nil {
			panic(err)
		}
		log.SetFlags(0) // disables default timestamping
		log.SetOutput(stream)
		log.SetPrefix("")
	} else {
		log.SetFlags(log.Ldate | log.Ltime)
		log.SetOutput(os.Stderr)
		log.SetPrefix(fmt.Sprintf("%s[%d] ", PROGRAM_NAME, syscall.Getpid()))
	}

	if command == "" {
		log.Fatalln("Command path is mandatory")
	}

	log.Println("Listening on", addr)
	sockfile, err := ListenFile(addr)
	if err != nil {
		log.Fatalln("Unable to open socket", err)
	}

	var targetUser *user.User
	if username != "" {
		targetUser, err = user.Lookup(username)
		if err != nil {
			log.Fatalln("Unable to find user", err)
		}
	}

	// Run the first process
	processGroup := MakeProcessGroup(*inputs, sockfile, targetUser)
	_, err = processGroup.StartProcess()
	if err != nil {
		log.Fatalln("Could not start process", err)
	}

	// Monitoring the processes
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR1)
	go handleSignals(processGroup, c, startTime, sockfile)

	// TODO: Full restart on USR2. Make sure the listener file is not set to SO_CLOEXEC
	// TODO: Restart processes if they die

	// For now, exit if no processes are left
	processGroup.WaitAll()
}
