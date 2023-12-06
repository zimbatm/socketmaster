package main

import "testing"

func Test_EnvironmentToMap(t *testing.T) {
	m := EnvironmentToMap([]string{"a=b=c"})
	if len(m) != 1 {
		t.Fatal("len")
	}
	if m["a"] != "b=c" {
		t.Fatal("a != b=c")
	}
}

func Test_MapToEnvironment(t *testing.T) {
	m := MapToEnvironment(map[string]string{"a": "b=c"})
	if len(m) != 1 {
		t.Fatal("len")
	}
	if m[0] != "a=b=c" {
		t.Fatal("a=b=c")
	}
}
