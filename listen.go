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
	case "unix": //, "unixpacket", "unixgram":
		var laddr *net.UnixAddr
		var listener *net.UnixListener

		laddr, err = net.ResolveUnixAddr(u.Scheme, u.Path)
		if err != nil {
			return
		}

		listener, err = net.ListenUnix(u.Scheme, laddr)
		if err != nil {
			return
		}

		file, err = listener.File()
	case "tcp", "tcp4", "tcp6":
		var laddr *net.TCPAddr
		var listener *net.TCPListener

		laddr, err = net.ResolveTCPAddr(u.Scheme, u.Host)
		if err != nil {
			return
		}

		listener, err = net.ListenTCP(u.Scheme, laddr)
		if err != nil {
			return
		}

		// Closing the listener doesn't affect the file and reversely.
		// http://golang.org/pkg/net/#TCPListener.File
		file, err = listener.File()
	default:
		err = fmt.Errorf("Unsupported scheme: %s", u.Scheme)
	}

	return
}
