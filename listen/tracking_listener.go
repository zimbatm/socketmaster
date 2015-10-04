package listen

import (
	"net"
	"sync"
)

type trackedConn struct {
	net.Conn
	listener *TrackingListener
	once     sync.Once
}

func (self *trackedConn) Close() error {
	// Makes sure Done() is not called twice here
	self.once.Do(func() { self.listener.wg.Done() })
	return self.Conn.Close()
}

/**
 * Keeps track of active connections. The WaitForChildren() function
 * can be used to only return when all child connections have been closed.
 */
type TrackingListener struct {
	net.Listener
	wg sync.WaitGroup
}

func NewTrackingListener(listener net.Listener) *TrackingListener {
	return &TrackingListener{
		Listener: listener,
	}
}

func (self *TrackingListener) Accept() (conn net.Conn, err error) {
	self.wg.Add(1)

	conn, err = self.Listener.Accept()
	if err != nil {
		self.wg.Done()
		return nil, err
	}

	conn = &trackedConn{
		Conn:     conn,
		listener: self,
	}

	return
}

func (self *TrackingListener) WaitForChildren() {
	self.wg.Wait()
}
