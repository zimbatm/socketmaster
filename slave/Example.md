```
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/zimbatm/socketmaster/slave"
)

var server *http.Server

type SleepyHandler struct{}

func (*SleepyHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.Println("New request")
	time.Sleep(time.Duration(10) * time.Second)
	log.Println("After sleep")
	res.WriteHeader(200)
}

func main() {
	var addr string
	flag.StringVar(&addr, "listen", "tcp://:8080", "Port to listen to")
	flag.Parse()

	server := &http.Server{
		Addr:    addr,
		Handler: new(SleepyHandler),
	}

	slave.Serve(server)

	log.Println("Bye bye")
}
```
