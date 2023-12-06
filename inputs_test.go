package main

import (
	"testing"
)

func Test_ParseInputs_Empty(t *testing.T) {
	inputs, err := ParseInputs([]string{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if inputs.commandlineConfig.Command != "" {
		t.Fatalf("wrong command value '%s'", inputs.commandlineConfig.Command)
	}
	if inputs.configFile != "" {
		t.Fatalf("wrong configFile value '%s'", inputs.configFile)
	}
}

func Test_ParseInputs_Command(t *testing.T) {
	inputs, err := ParseInputs([]string{"--command", "hello"})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if inputs.commandlineConfig.Command != "hello" {
		t.Fatalf("wrong command value '%s'", inputs.commandlineConfig.Command)
	}
}

func Test_ParseInputs_ConfigFile(t *testing.T) {
	inputs, err := ParseInputs([]string{"--config-file", "testdata/command.yaml"})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if inputs.configFile != "testdata/command.yaml" {
		t.Fatalf("wrong command value '%s'", inputs.commandlineConfig.Command)
	}
}

func Test_ParseInputs_ConfigFile_Command_Error(t *testing.T) {
	inputs, err := ParseInputs([]string{"--config-file", "testdata/command.yaml", "--command", "would-be-forgotten"})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	unused, err := inputs.LoadConfig()

	if unused != nil {
		t.Fatal("LoadConfig must fail")
	}

	if err.Error() != "Between the command line and the config file 'testdata/command.yaml', command can only be set once." {
		t.Fatalf("Wrong error, was '%s'", err.Error())
	}

}
