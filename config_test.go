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

// environment
func Test_Config_Environment(t *testing.T) {
	var config Config
	err := yaml.Unmarshal([]byte("command: hi\nenvironment:\n  hello: world"), &config)
	// err := config.LoadString(`command: hi`)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if config.Environment["hello"] != "world" {
		t.Fatalf("wrong environment.hello value '%s'", config.Environment["hello"])
	}
}

func Test_Config_Merge_Environment(t *testing.T) {
	a := Config{Environment: map[string]string{}}
	b := Config{Environment: map[string]string{"foo": "bar"}}

	a.Merge(b)

	if a.Environment["foo"] != "bar" {
		t.Fatal("foo isn't bar")
	}
}
