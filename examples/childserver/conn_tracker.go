package main

import (
	"net"
	"sync"
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

func NewTrackingListener(listener net.Listener) *TrackingListener {
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
