/**
 * Utility package that counts the number of active connections.
 *
 * Useful to gracefully shutdown your webserver.
 */
package socketmaster

import (
	"net/http"
	"sync"
)

type ConnectionCountHandler struct {
	parent     http.Handler
	childCount sync.WaitGroup
}

func NewConnectionCountHandler(parent http.Handler, childCount sync.WaitGroup) *ConnectionCountHandler {
	return &ConnectionCountHandler{
		parent:     parent,
		childCount: childCount,
	}
}

func (self *ConnectionCountHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	self.childCount.Add(1)
	defer func() {
		self.childCount.Done()
	}()
	self.parent.ServeHTTP(res, req)
}
