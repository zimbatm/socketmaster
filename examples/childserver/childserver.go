package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var server *http.Server
var fd uintptr
var listener net.Listener
var childCount sync.WaitGroup

type SleepyHandler struct{}

func (*SleepyHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.Println("New request")
	time.Sleep(time.Duration(10) * time.Second)
	log.Println("After sleep")
	res.WriteHeader(200)
}

func init() {
	sleepyHandler := new(SleepyHandler)
	connHandler := NewConnectionCountHandler(sleepyHandler, childCount)

	server = &http.Server{
		Handler: connHandler,
	}
}

func main() {
	var addr string
	flag.StringVar(&addr, "listen", "tcp://:8080", "Port to listen to")
	flag.Parse()

	// Transform fd into listener
	listener, err := Listen(addr)
	if err != nil {
		log.Fatalln("Unable to open FD", fd, err)
	}

	log.Println("Starting web server on", addr)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func(c chan os.Signal, listener net.Listener) {
		for {
			<-c // os.Signal
			log.Println("Closing listener")
			listener.Close()
		}
	}(c, listener)

	server.Serve(listener)

	log.Println("Waiting for children to close")

	childCount.Wait()

	log.Println("Bye bye")
}
