package socketchild

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type trackedConn struct {
	net.Conn
	listener *TrackingListener
	once     sync.Once
}

// TODO: Is Close called even if it's the client who closed the connection ?
func (self *trackedConn) Close() error {
	// Make sure Done() is not called twice here
	self.once.Do(func() { self.listener.wg.Done() })
	return self.Conn.Close()
}

type TrackingListener struct {
	net.Listener
	wg sync.WaitGroup
}

func newTrackingListener(listener net.Listener) *TrackingListener {
	return &TrackingListener{
		Listener: listener,
	}
}

func (self *TrackingListener) Accept() (net.Conn, error) {
	self.wg.Add(1)

	conn, err := self.Listener.Accept()
	if err != nil {
		self.wg.Done()
		return nil, err
	}

	conn2 := &trackedConn{
		Conn:     conn,
		listener: self,
	}

	return conn2, nil
}

func (self *TrackingListener) WaitForChildren() {
	self.wg.Wait()
}

//opens a Listener on file descriptor 3.
func Listen() (listener net.Listener, err error) {
	sockfile := os.NewFile(uintptr(3), fmt.Sprintf("fd://%d", 3))
	listener, err = net.FileListener(sockfile)
	return
}

func GetTrackingListener() (*TrackingListener, error) {
	listener, err := Listen()
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	trackerListener := newTrackingListener(listener)
	return trackerListener, nil
}

func SetupSignalHandling(trackerListener *TrackingListener) {
	f := func() {
		log.Println("Closing listener")
		if err := trackerListener.Close(); err != nil {
			log.Printf("error closing trackerListener: %v\n", err)
		}
	}
	handleSignal(f)
}

func SetupProcessTerminationHandling() {
	handleSignal(func() { os.Exit(0) })
}

func handleSignal(f func()) {
	c := make(chan os.Signal, 10)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR2)
	go func() {
		for {
			sig := <-c // os.Signal
			log.Printf("socketchild captured %v\n", sig)
			f()
		}
	}()
}

//opens a Listener on file descriptor 3. serves http on it until INT, TERM, QUIT or HUP signals are retrieved.
//  Waits for at most 30 seconds for responses to finish.
func ListenAndServe(handler http.Handler, trackerListener *TrackingListener) (err error) {
	if handler == nil {
		handler = http.DefaultServeMux
	}

	server := &http.Server{Addr: "fd://3", Handler: handler}

	server.Serve(trackerListener)

	childrenQuit := make(chan bool)
	go func() {
		log.Println("Waiting for children to close")
		trackerListener.WaitForChildren()
		childrenQuit <- true
	}()

	select {
	case <-childrenQuit:
		log.Printf("Children all quit")
	case <-time.After(30 * time.Second):
		log.Printf("Timed out. returning immediately.")
	}

	return
}

func GetSocketMasterProcess() *os.Process {
	smPid := os.Getenv("SOCKETMASTER_PID")
	if smPid == "" {
		log.Fatalln("Cannot get env SOCKETMASTER_PID")
	}

	pid, err := strconv.Atoi(smPid)
	if err != nil {
		log.Fatalln("Cannot get socketmaster PID")
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		log.Fatalln("Cannot get socketmaster process, PID: %v", pid)
	}

	return p
}

func SendReadySignalToSocketMaster(p *os.Process) {
	signalStr := os.Getenv("CHILD_READY_SIGNAL")
	if signalStr == "" {
		return
	}

	signal, err := strconv.Atoi(signalStr)
	if err != nil {
		log.Fatalln("Cannot get env CHILD_READY_SIGNAL")
	}

	if signal == 0 {
		return
	}

	err = p.Signal(syscall.Signal(signal))
	if err != nil {
		log.Fatalln("Cannot send signal %d to socketmaster %d", signal, p.Pid)
	}
}
