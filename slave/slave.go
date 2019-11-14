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

func ListenAndServeHTTP(server *http.Server, address string, timeout time.Duration) error {
	l, err := Listen(address)
	if err != nil {
		return err
	}

	signalHandler(func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		server.Shutdown(ctx)
	})

	// Start serving.
	server.Addr = address
	return server.Serve(l)
}

func ListenAndServeGRPC(server *grpc.Server, address string, timeout time.Duration) error {
	l, err := Listen(address)
	if err != nil {
		return err
	}

	signalHandler(func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		stopped := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(stopped)
		}()

		select {
		case <-ctx.Done():
		case <-stopped:
		}
	})

	// Start serving.
	return server.Serve(l)
}

func signalHandler(callbackFn func()) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		<-c
		log.Println("Shutting down gracefully the server")
		callbackFn()
		log.Println("The server did shut down")
	}()
}
