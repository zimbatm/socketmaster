package socketchild

import (
  "net"
  "sync"
  "os"
  "fmt"
  "os/signal"
  "net/http"
  "log"
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
func Listen() (listener net.Listener, err error){
  sockfile := os.NewFile(uintptr(3), fmt.Sprintf("fd://%d", 3))
  listener, err = net.FileListener(sockfile)
  return
}

//opens a Listener on file descriptor 3. serves http on it until INT, TERM, QUIT or HUP signals are retrieved.
//  Waits for at most 30 seconds for responses to finish.
func ListenAndServe(handler http.Handler) (err error){
  if handler == nil {
    handler = http.DefaultServeMux
  }

  server := &http.Server{Addr: "fd://3", Handler: handler}

  listener, err := Listen()
  if err != nil {
    log.Fatalln(err)
    return
  }

  trackerListener := newTrackingListener(listener)

  c := make(chan os.Signal)
  signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
  go func() {
    for {
      <-c // os.Signal
      log.Println("Closing listener")
      trackerListener.Close()
    }
  }()

  server.Serve(trackerListener)

  childrenQuit := make(chan bool)
  go func(){
    log.Println("Waiting for children to close")
    trackerListener.WaitForChildren()
    childrenQuit <- true
  }()

  select{
  case <-childrenQuit:
    log.Printf("Children all quit")
  case <-time.After(30 * time.Second):
    log.Printf("Timed out. returning immediately.")
  }

  return
}