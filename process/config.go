package process

import (
	"os"
	"os/user"
)

type Config struct {
	command string
	files   []*os.File
	user    *user.User
}

func NewConfig(command string, files []*os.File, u *user.User) *Config {
	return &Config{command, files, u}
}
