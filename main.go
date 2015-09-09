package main

import (
	"flag"
	"fmt"
	"github.com/zimbatm/socketmaster/listen"
	"github.com/zimbatm/socketmaster/process"
	"github.com/zimbatm/socketmaster/sd_daemon"
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

func handleSignals(group *process.Group, config *process.Config, c <-chan os.Signal, startTime int) {
	for {
		signal := <-c // os.Signal
		syscallSignal := signal.(syscall.Signal)

		switch syscallSignal {
		case syscall.SIGHUP:
			p, err := group.Start(config)
			if err != nil {
				log.Printf("Could not start new process: %v\n", err)
			} else {
				if startTime > 0 {
					time.Sleep(time.Duration(startTime) * time.Millisecond)
				}

				// A possible improvement woud be to only swap the
				// process if the new child is still alive.
				group.SignalAll(syscall.SIGTERM, p)
			}
		default:
			// Forward signal
			group.SignalAll(signal, nil)
		}
	}
}

func main() {
	var (
		listen_      *StrValues = NewStrValues()
		command      string
		err          error
		notifySocket string
		startTime    int
		useSyslog    bool
		username     string
	)

	flag.StringVar(&command, "command", "", "Program to start")
	flag.Var(listen_, "listen", "Port on which to bind (eg: :8080). Can be invoked multiple times.")
	flag.StringVar(&notifySocket, "notify_socket", "", "If sets waits for the child to notify on that socket")
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
		log.SetOutput(os.Stdout)
		log.SetPrefix(fmt.Sprintf("%s[%d] ", PROGRAM_NAME, syscall.Getpid()))
	}

	if command == "" {
		log.Fatalln("Missing command path")
	}

	if command, err = exec.LookPath(command); err != nil {
		log.Fatalln("Could not find executable", err)
	}

	var targetUser *user.User
	if username != "" {
		targetUser, err = user.Lookup(username)
		if err != nil {
			log.Fatalln("Unable to find user", err)
		}
	}

	files := sd_daemon.ListenFds(true)
	if files != nil && len(files) > 0 {
		log.Printf("Inherited %d files", len(files))
	} else {
		addrs := listen_.Value()
		files = make([]*os.File, len(addrs))
		for i, addr := range addrs {
			log.Printf("Listening on %s", addr)
			file, _, err := listen.Listen(addr)
			if err != nil {
				log.Fatal(err)
			}
			files[i] = file
		}
	}

	config := process.NewConfig(command, files, targetUser)

	// Run the first process
	group := process.NewGroup()
	_, err = group.Start(config)
	if err != nil {
		log.Fatalln("Could not start process", err)
	}

	// Monitoring the processes
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go handleSignals(group, config, c, startTime)

	// TODO: Full restart on USR2. Make sure the listener file is not set to SO_CLOEXEC

	// For now, exit if no processes are left
	group.WaitAll()
}
