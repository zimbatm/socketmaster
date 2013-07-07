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

const PROGRAM_NAME = "socketmaster"

func handleSignals(processGroup *ProcessGroup, c <-chan os.Signal, startTime int) {
	for {
		signal := <-c // os.Signal
		syscallSignal := signal.(syscall.Signal)

		switch syscallSignal {
		case syscall.SIGHUP:
			process, err := processGroup.StartProcess()
			if err != nil {
				log.Printf("Could not start new process: %v\n", err)
			} else {
				if startTime > 0 {
					time.Sleep(time.Duration(startTime) * time.Millisecond)
				}

				// A possible improvement woud be to only swap the
				// process if the new child is still alive.
				processGroup.SignalAll(signal, process)
			}
		default:
			// Forward signal
			processGroup.SignalAll(signal, nil)
		}
	}
}

func main() {
	var (
		addr      string
		command   string
		err       error
		startTime int
		useSyslog bool
		username  string
	)

	flag.StringVar(&command, "command", "", "Program to start")
	flag.StringVar(&addr, "listen", "tcp://:8080", "Port on which to bind")
	flag.IntVar(&startTime, "start", 3000, "How long the new process takes to boot in millis")
	flag.BoolVar(&useSyslog, "syslog", false, "Log to syslog")
	flag.StringVar(&username, "user", "", "run the command as this user")
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
		log.SetFlags(log.Ldate | log.Ltime)
		log.SetOutput(os.Stderr)
		log.SetPrefix(fmt.Sprintf("%s[%d] ", PROGRAM_NAME, syscall.Getpid()))
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

	// Run the first process
	processGroup := MakeProcessGroup(commandPath, sockfile, targetUser)
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
