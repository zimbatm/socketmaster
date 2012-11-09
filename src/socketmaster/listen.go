// Utility to listen on tcp[46]? sockets or a file descriptor
package socketmaster

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

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

// Almost the same as Listen but returns the undelying File object instead
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
