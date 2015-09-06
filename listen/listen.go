package listen

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

/**
 * Like `net.Listen` but more flexible. Returns both the listener and the
 * file-descriptor since the net.Listener interface doesn't allow to extract
 * the File().
 *
 * Example inputs:
 *   :8080
 *   fd://3
 *   tcp://:3000
 *   tcp4://localhost:3000
 *   tcp6://:3000
 *   unix:///var/run/socketmaster.sock
 *
 * Go automatically opens ports with SO_REUSEADDR where it makes sense.
 */
func Listen(addr string) (file *os.File, l net.Listener, err error) {
	if !strings.Contains(addr, "://") {
		addr = "tcp://" + addr
	}

	u, err := url.Parse(addr)
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
		l, err = net.FileListener(file)
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

		l = net.Listener(listener)
		file, err = listener.File()
	case "tcp", "tcp4", "tcp6", "":
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

		l = net.Listener(listener)
		// Closing the listener doesn't affect the file and reversely.
		// http://golang.org/pkg/net/#TCPListener.File
		file, err = listener.File()
	default:
		err = fmt.Errorf("Unsupported scheme: %s", u.Scheme)
	}

	return
}
