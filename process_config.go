package main

import (
	"os"
	"os/user"
)

type ProcessConfig struct {
	command string
	files   []*os.File
	user    *user.User
}

func NewProcessConfig(command string, files []*os.File, u *user.User) *ProcessConfig {
	return &ProcessConfig{command, files, u}
}
