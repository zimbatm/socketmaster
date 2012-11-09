package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

// Utility to listen on tcp[46]? sockets or a file descriptor
func Listen(rawurl string) (listener net.Listener, err error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return
	}

	switch u.Scheme {
	case "fd":
		var fd uint64
		fd, err = strconv.ParseUint(u.Host, 10, 8)
		if err != nil {
			return
		}
		// NOTE: The name argument doesn't really matter apparently
		sockfile := os.NewFile(uintptr(fd), fmt.Sprintf("fd://%d", fd))
		listener, err = net.FileListener(sockfile)
		if err != nil {
			return
		}
	default:
		var laddr *net.TCPAddr
		laddr, err = net.ResolveTCPAddr(u.Scheme, u.Host)
		if err != nil {
			return
		}

		listener, err = net.ListenTCP("tcp", laddr)
		if err != nil {
			return
		}
	}

	return
}
