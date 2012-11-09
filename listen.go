package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

// Utility to open a tcp[46]? or fd
func ListenFile(rawurl string) (file *os.File, err error) {
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
		file = os.NewFile(uintptr(fd), fmt.Sprintf("fd://%d", fd))
	default:
		var laddr *net.TCPAddr
		var listener *net.TCPListener
		laddr, err = net.ResolveTCPAddr(u.Scheme, u.Host)
		if err != nil {
			return
		}

		listener, err = net.ListenTCP("tcp", laddr)
		if err != nil {
			return
		}

		// TODO: Is the listener going to close the file when garbage-collected ?
		file, err = listener.File()
	}

	return
}
