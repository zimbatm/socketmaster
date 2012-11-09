package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"socketmaster"
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
	connHandler := socketmaster.NewConnectionCountHandler(sleepyHandler, childCount)

	server = &http.Server{
		Handler: connHandler,
	}
}

func main() {
	// Transform fd into listener
	listener, err := socketmaster.Listen("fd://3")
	if err != nil {
		log.Fatalln("Unable to open FD", fd, err)
	}

	log.Println("Starting web server")

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
