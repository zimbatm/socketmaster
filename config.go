package main

import (
	"errors"
	"flag"
	"io/ioutil"

	yaml "gopkg.in/yaml.v3"
)

/*
	flag.StringVar(&command, "command", "", "Program to start")
	flag.StringVar(&addr, "listen", "tcp://:8080", "Port on which to bind")
	flag.IntVar(&startTime, "start", 3000, "How long the new process takes to boot in millis")
	flag.BoolVar(&useSyslog, "syslog", false, "Log to syslog")
	flag.StringVar(&username, "user", "", "run the command as this user")
*/

// The mutable configuration items
type Config struct {
	Command string `yaml:"command"`
}

func emptyConfig() Config {
	return Config{
		Command: "",
	}
}

func (config *Config) LoadFile(path string) error {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return config.LoadBytes(yamlFile)
}

func (config *Config) LoadBytes(yamlString []byte) error {
	return yaml.Unmarshal(yamlString, &config)
}

func (config *Config) LoadString(yamlString string) error {
	return config.LoadBytes([]byte(yamlString))
}

// Some of the mutable flags can be set on the command line as well.
func (config *Config) LinkToFlagSet(flags *flag.FlagSet) {
	flags.StringVar(&config.Command, "command", config.Command, "Program to start")
}

func (config *Config) Merge(other Config) error {
	empty := emptyConfig()

	if other.Command != empty.Command {
		if config.Command == empty.Command {
			config.Command = other.Command
		} else {
			return errors.New("command can only be set once.")
		}
	}

	return nil
}
