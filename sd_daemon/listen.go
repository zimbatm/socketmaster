package sd_daemon

import (
	"os"
	"strconv"
	"syscall"
)

const LISTEN_FDS_START = 3

// See http://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
func ListenFds(unsetEnv bool) []*os.File {
	if unsetEnv {
		defer os.Unsetenv("LISTEN_PID")
		defer os.Unsetenv("LISTEN_FDS")
	}

	pid, err := strconv.Atoi(os.Getenv("LISTEN_PID"))
	if err != nil || pid != os.Getpid() {
		return nil
	}

	nfds, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
	if err != nil || nfds <= 0 {
		return nil
	}

	files := make([]*os.File, 0, nfds)
	for fd := LISTEN_FDS_START; fd < LISTEN_FDS_START+nfds; fd++ {
		syscall.CloseOnExec(fd)
		files = append(files, os.NewFile(uintptr(fd), "LISTEN_FD_"+strconv.Itoa(fd)))
	}

	return files
}
