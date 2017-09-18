package slave

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net"
	"net/http"
)

type app struct {
	server   *http.Server
	listener net.Listener
}

func newApp(server *http.Server) *app {
	return &app{
		server: server,
	}
}
func (a *app) Listener() net.Listener {
	return a.listener
}

func (a *app) serve() {
	a.server.Serve(a.listener)
}

func (a *app) listen() error {
	l, err := Listen(a.server.Addr)
	if err != nil {
		return err
	}

	a.listener = l
	return nil
}

func (a *app) signalHandler(d time.Duration) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), d)
		defer cancel()

		log.Println("Shutting down gracefully the server")
		a.server.Shutdown(ctx)
		log.Println("The server did shut down")
	}()
}

func Serve(server *http.Server, timeout time.Duration) error {
	a := newApp(server)

	// Acquire Listeners
	if err := a.listen(); err != nil {
		return err
	}

	a.signalHandler(timeout)

	// Start serving.
	a.serve()

	return nil
}
