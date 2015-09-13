package main

import (
	"flag"
	"fmt"
	"github.com/zimbatm/socketmaster/listen"
	"github.com/zimbatm/socketmaster/process"
	"github.com/zimbatm/socketmaster/sd_daemon"
	"log"
	"log/syslog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"
)

const PROGRAM_NAME = "socketmaster"

func main() {
	var (
		command      string
		dir          string
		err          error
		listen_      *StrValues = NewStrValues()
		logger       *log.Logger
		notifyAddr   string
		startTimeout int
		stopTimeout  int
		useSyslog    bool
		username     string
	)

	flag.StringVar(&command, "cmd", "", "Program to start")
	flag.StringVar(&dir, "dir", "", "Work directory of the command")
	flag.Var(listen_, "listen", "Port on which to bind (eg: :8080). Can be invoked multiple times.")
	flag.StringVar(&notifyAddr, "notify", "", "Path to the notification socket.")
	flag.IntVar(&startTimeout, "start", 0, "Maximum time for the new process to boot (millis)")
	flag.IntVar(&stopTimeout, "stop", 0, "Maximum time for the old processes to stop (millis)")
	flag.BoolVar(&useSyslog, "syslog", false, "Log to syslog")
	flag.StringVar(&username, "user", "", "run the command as this user")
	flag.Parse()

	if useSyslog {
		stream, err := syslog.New(syslog.LOG_INFO, PROGRAM_NAME)
		if err != nil {
			log.Fatal(err)
		}
		logger = log.New(stream, "", 0)
	} else {
		prefix := fmt.Sprintf("%s[%d] ", PROGRAM_NAME, syscall.Getpid())
		logger = log.New(os.Stdout, prefix, log.Ldate|log.Ltime)
	}

	if command == "" {
		logger.Fatalln("Missing command path")
	}

	if command, err = exec.LookPath(command); err != nil {
		logger.Fatalln("Could not find executable", err)
	}

	var targetUser *user.User
	if username != "" {
		targetUser, err = user.Lookup(username)
		if err != nil {
			logger.Fatalln("Unable to find user", err)
		}
	}

	files := sd_daemon.ListenFds(true)
	if files != nil && len(files) > 0 {
		logger.Printf("Inherited %d files", len(files))
	} else {
		addrs := listen_.Value()
		files = make([]*os.File, len(addrs))
		for i, addr := range addrs {
			logger.Printf("Listening on %s", addr)
			file, _, err := listen.Listen(addr)
			if err != nil {
				logger.Fatal(err)
			}
			files[i] = file
		}
	}

	var notifyConn *net.UnixConn
	if notifyAddr != "" {
		addr := net.UnixAddr{
			Name: notifyAddr,
			Net:  "unixgram",
		}
		notifyConn, err = net.ListenUnixgram(addr.Net, &addr)
		if err != nil {
			logger.Fatal(err)
		}
	}

	config := &process.Config{
		Command:      command,
		Dir:          dir,
		Files:        files,
		NotifyConn:   notifyConn,
		StartTimeout: time.Millisecond * time.Duration(startTimeout),
		StopTimeout:  time.Millisecond * time.Duration(stopTimeout),
		User:         targetUser,
	}

	ready := make(chan bool)
	go func(ready <-chan bool) {
		<-ready
		sd_daemon.NotifyReady()
	}(ready)

	signals := make(chan os.Signal)
	signal.Notify(signals)

	manager := process.StartManager(config, logger, signals, ready)

	// TODO: Full restart on USR2
	err = manager.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
