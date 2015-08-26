package slave

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"net"
	"net/http"
)

type app struct {
	server   *http.Server
	listener *TrackingListener
}

func newApp(server *http.Server) *app {
	return &app{
		server: server,
	}
}
func (a *app) Listener() net.Listener {
	return a.listener
}

func (a *app) wait() {
	a.listener.WaitForChildren()
}

func (a *app) serve() {
	a.server.Serve(a.listener)
}

func (a *app) listen() error {
	l, err := Listen(a.server.Addr)
	if err != nil {
		return err
	}

	a.listener = NewTrackingListener(l)
	return nil
}

func (a *app) signalHandler() {

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func(c chan os.Signal, listener net.Listener) {
		<-c
		log.Println("Closing listener")
		listener.Close()
	}(c, a.listener)
}

func Serve(server *http.Server) error {
	a := newApp(server)

	// Acquire Listeners
	if err := a.listen(); err != nil {
		return err
	}

	a.signalHandler()

	// Start serving.
	a.serve()

	log.Println("Waiting for children to close")

	a.wait()

	return nil
}
