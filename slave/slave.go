package slave

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

type httpAPP struct {
	server *http.Server
}

func (a *httpAPP) ShutdownFunc(timeout time.Duration) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		a.server.Shutdown(ctx)
	}
}

type grpcAPP struct {
	server *grpc.Server
}

func (a *grpcAPP) ShutdownFunc(timeout time.Duration) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		stopped := make(chan struct{})
		go func() {
			a.server.GracefulStop()
			close(stopped)
		}()

		select {
		case <-ctx.Done():
		case <-stopped:
		}
	}
}

func ListenAndServeHTTP(server *http.Server, address string, timeout time.Duration) error {
	server.Addr = address

	app := &httpAPP{server: server}

	l, err := Listen(server.Addr)
	if err != nil {
		return err
	}

	signalHandler(app.ShutdownFunc(timeout))

	// Start serving.
	return app.server.Serve(l)
}

func ListenAndServeGRPC(server *grpc.Server, address string, timeout time.Duration) error {
	app := &grpcAPP{server: server}

	l, err := Listen(address)
	if err != nil {
		return err
	}

	signalHandler(app.ShutdownFunc(timeout))

	// Start serving.
	return app.server.Serve(l)
}

func signalHandler(shutdownFn func()) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		<-c
		log.Println("Shutting down gracefully the server")
		shutdownFn()
		log.Println("The server did shut down")
	}()
}
