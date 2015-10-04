package main

import (
	"flag"
	"github.com/zimbatm/socketmaster/listen"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var server *http.Server

type SleepyHandler struct{}

func (*SleepyHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.Println("New request")
	time.Sleep(time.Duration(10) * time.Second)
	log.Println("After sleep")
	res.WriteHeader(200)
	res.Write([]byte("Hi\n"))
}

func init() {
	sleepyHandler := new(SleepyHandler)

	server = &http.Server{
		Handler: sleepyHandler,
	}
}

func main() {
	var addr string
	flag.StringVar(&addr, "listen", "tcp://:8080", "Port to listen to")
	flag.Parse()

	// Transform fd into listener
	fs, err := listen.ListenFiles([]string{addr})
	if err != nil {
		log.Fatalln(err)
	}

	listener, err := net.FileListener(fs[0])
	if err != nil {
		log.Fatalln(err)
	}
	f.Close()

	trackerListener := listen.NewTrackingListener(listener)

	log.Println("Starting web server on", addr)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func(c chan os.Signal, listener net.Listener) {
		for {
			<-c // os.Signal
			log.Println("Closing listener")
			trackerListener.Close()
		}
	}(c, trackerListener)

	err = server.Serve(trackerListener)

	log.Println("Waiting for children to close", err)

	trackerListener.WaitForChildren()

	log.Println("Bye bye")
}
