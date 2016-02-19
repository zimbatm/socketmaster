package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"
)

const (
	PROGRAM_NAME = "socketmaster"
	SIGUSR1_NAME = "SIGUSR1"
)

//Dir to use to start children. Automatically inferred from os.Args[0].
//We need this in order to support symlink-based deploys where we use the launch path of socketmaster to track the dir instead of putting the full path in the command arg.
// $> /some/path/symlink/bin/socketmaster -command=hi
// if you change what 'symlink' points to, then you want to launch the *new* version
var WorkingDir string

func init() {
	WorkingDir = os.Getenv("APP_HOME")
	fmt.Printf("Setting Working Directory to: %#v\n", WorkingDir)
}

type SignalAllParams struct {
	signal os.Signal
}

// TODO:
// should use goroutines for each `case`
// so that master process will not lose upcoming signals
func handleSignals(processGroup *ProcessGroup, c <-chan os.Signal, startTime int, childReadySignal syscall.Signal) {
	cachedSignals := []SignalAllParams{}

	for {
		signal := <-c // os.Signal
		syscallSignal := signal.(syscall.Signal)

		log.Printf("socketmaster captured %v\n", syscallSignal)

		switch syscallSignal {
		case syscall.SIGHUP, syscall.SIGUSR2:
			_, err := processGroup.StartProcess(childReadySignal)
			if err != nil {
				log.Printf("Could not start new process: %v\n", err)
			} else {
				if startTime > 0 {
					time.Sleep(time.Duration(startTime) * time.Millisecond)
				}
				// childReadySignal == 0 means socketmaster do not need to wait for child ready
				if childReadySignal == syscall.Signal(0) {
					// A possible improvement woud be to only swap the
					// process if the new child is still alive.
					processGroup.SignalAll(signal, processGroup.LastProcess)
				} else {
					cachedSignals = append(cachedSignals, SignalAllParams{signal})
				}
			}
		case childReadySignal:
			if len(cachedSignals) > 0 {
				params := cachedSignals[0]
				cachedSignals = cachedSignals[1:]
				processGroup.SignalAll(params.signal, processGroup.LastProcess)
				// A possible improvement woud be to only swap the
				// process if the new child is still alive.
				//processGroup.SignalAll(syscall.SIGTERM, process)
			}
		default:
			// Forward signal
			processGroup.SignalAll(signal, nil)
		}
	}
}

func main() {
	var (
		addr                string
		command             string
		err                 error
		startTime           int
		useSyslog           bool
		username            string
		childReadySignalStr string
		childReadySignal    syscall.Signal
	)

	flag.StringVar(&command, "command", "", "Program to start")
	flag.StringVar(&addr, "listen", "tcp://:8080", "Port on which to bind")
	flag.IntVar(&startTime, "start", 3000, "How long the new process takes to boot in millis")
	flag.BoolVar(&useSyslog, "syslog", false, "Log to syslog")
	flag.StringVar(&username, "user", "", "run the command as this user")
	flag.StringVar(&childReadySignalStr, "child_ready_signal", "", "The signal child process will send when it's ready")
	flag.Parse()

	if useSyslog {
		stream, err := syslog.New(syslog.LOG_INFO, PROGRAM_NAME)
		if err != nil {
			panic(err)
		}
		log.SetFlags(0) // disables default timestamping
		log.SetOutput(stream)
		log.SetPrefix("")
	} else {
		log.SetFlags(0)
		log.SetOutput(os.Stderr)
		log.SetPrefix(fmt.Sprintf("[%d] ", syscall.Getpid()))
	}

	if command == "" {
		log.Fatalln("Command path is mandatory")
	}

	commandPath, err := exec.LookPath(command)
	if err != nil {
		log.Fatalln("Could not find executable", err)
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

	switch childReadySignalStr {
	case SIGUSR1_NAME:
		childReadySignal = syscall.SIGUSR1
	case "":
		childReadySignal = syscall.Signal(0)
	default:
		log.Fatalln("Do not support ", childReadySignalStr)
	}

	// Run the first process
	processGroup := MakeProcessGroup(commandPath, sockfile, targetUser)
	_, err = processGroup.StartProcess(childReadySignal)
	if err != nil {
		log.Fatalln("Could not start process", err)
	}

	// Monitoring the processes
	// NOTE: add buffer because `handleSignals` is not fast enough
	// TODO: rewrite `handleSignals` so signals will not be lost
	c := make(chan os.Signal, 1)
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR2}
	if childReadySignal != syscall.Signal(0) {
		signals = append(signals, childReadySignal)
	}
	signal.Notify(c, signals...)
	go handleSignals(processGroup, c, startTime, childReadySignal)

	// TODO: Full restart on USR2. Make sure the listener file is not set to SO_CLOEXEC
	// TODO: Restart processes if they die

	// For now, exit if no processes are left
	processGroup.WaitAll()
}
