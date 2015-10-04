package listen

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

const LISTEN_FDS_START = 3

var listen_fds = 0

/**
 * Like ListenFile but also uses the systemd socket-activation when available.
 *
 * See http://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
 *
 * To match the ordering the user has to pass the same number of
 * addresses to the function as the LISTEN_FDS.
 *
 * This function is not thread-safe.
 */
func ListenFiles(addrs []string) (files []*os.File, err error) {
	loadListenFds()
	if listen_fds > 0 { // Inherit from socket activation
		if len(addrs) != listen_fds {
			err = fmt.Errorf("Doesn't have a matching number of LISTEN_FDS: %d != %d", len(addrs), listen_fds)
			return
		}

		files = make([]*os.File, listen_fds)
		for i := 0; i < listen_fds; i++ {
			fd := i + LISTEN_FDS_START
			syscall.CloseOnExec(fd)
			files[i] = os.NewFile(uintptr(fd), addrs[i])
		}
	} else {
		files = make([]*os.File, len(addrs))
		for i, addr := range addrs {
			files[i], err = ListenFile(addr)
			if err != nil {
				return
			}
		}
	}

	return
}

// Load the number of fds at start and remove the environment variables
func loadListenFds() {
	if listen_fds > 0 {
		return
	}
	if fds := os.Getenv("LISTEN_FDS"); fds != "" {
		nfds, err := strconv.Atoi(fds)
		if err != nil || nfds <= 0 {
			return
		}

		pid, err := strconv.Atoi(os.Getenv("LISTEN_PID"))
		if err != nil || pid != os.Getpid() {
			return
		}

		os.Unsetenv("LISTEN_PID")
		os.Unsetenv("LISTEN_FDS")
		listen_fds = nfds
	}
}
