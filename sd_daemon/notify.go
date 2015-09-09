package sd_daemon

// Code lifted from https://github.com/coreos/go-systemd/blob/master/daemon/sdnotify.go

import (
	"errors"
	"net"
	"os"
)

var SdNotifyNoSocket = errors.New("No socket")

const NOTIFY_SOCKET = "NOTIFY_SOCKET"

// Notify sends a message to the init daemon. It is common to ignore the error.
func Notify(unsetEnv bool, state string) error {
	notifySocket := os.Getenv(NOTIFY_SOCKET)
	if notifySocket == "" {
		return SdNotifyNoSocket
	}
	if unsetEnv {
		defer os.Unsetenv(NOTIFY_SOCKET)
	}

	socketAddr := &net.UnixAddr{
		Name: notifySocket,
		Net:  "unixgram",
	}

	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(state))
	return err
}

func NotifyReady() error {
	return Notify(true, "READY=1")
}
