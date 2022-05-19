package main

import (
	"errors"
	"flag"
	"fmt"
)

type Inputs struct {
	commandlineConfig Config

	configFile string

	addr      string
	startTime int
	useSyslog bool
	username  string
}

func ParseInputs(args []string) (*Inputs, error) {
	var inputs Inputs
	flags := flag.NewFlagSet("socketmaster", flag.ExitOnError)

	inputs.commandlineConfig.LinkToFlagSet(flags)
	flags.StringVar(&inputs.configFile, "config-file", "", "Configuration file to load and watch")
	flags.StringVar(&inputs.addr, "listen", "tcp://:8080", "Port on which to bind")
	flags.IntVar(&inputs.startTime, "start", 3000, "How long the new process takes to boot in millis")
	flags.BoolVar(&inputs.useSyslog, "syslog", false, "Log to syslog")
	flags.StringVar(&inputs.username, "user", "", "run the command as this user")

	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}

	return &inputs, err
}

func (inputs *Inputs) LoadConfig() (*Config, error) {

	if inputs.configFile == "" {
		return &inputs.commandlineConfig, nil
	}

	config := inputs.commandlineConfig

	var fileConfig Config
	err := fileConfig.LoadFile(inputs.configFile)
	if err != nil {
		return nil, err
	}

	err = config.Merge(fileConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(
			"Between the command line and the config file '%s', %s", inputs.configFile, err.Error()))
	}

	return &config, nil
}
