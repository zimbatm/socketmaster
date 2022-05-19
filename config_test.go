package main

import (
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func Test_Config_Empty(t *testing.T) {
	var config Config
	err := config.LoadBytes(
		[]byte(""),
	)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if config.Command != "" {
		t.Fatalf("wrong Command value '%s'", config.Command)
	}
}

func Test_Config_Command(t *testing.T) {
	var config Config
	err := yaml.Unmarshal([]byte(`command: hi`), &config)
	// err := config.LoadString(`command: hi`)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if config.Command != "hi" {
		t.Fatalf("wrong command value '%s'", config.Command)
	}
}
