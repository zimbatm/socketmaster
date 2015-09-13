package process

import (
	"net"
	"os"
	"os/user"
	"time"
)

type Config struct {
	Command      string
	Dir          string
	Files        []*os.File
	NotifyConn   *net.UnixConn
	StartTimeout time.Duration
	StopTimeout  time.Duration
	User         *user.User
}
