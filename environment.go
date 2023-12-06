package main

import "strings"

func EnvironmentToMap(envStrings []string) map[string]string {
	envMap := make(map[string]string)
	for _, envLine := range envStrings {

		// Get the variable name
		var name string

		var silly = strings.Split(envLine, "=")
		if len(silly) == 0 {
			// pathological, skip
			continue
		}

		name = silly[0]

		value := envLine[len(name)+1:]

		envMap[name] = value
	}
	return envMap
}

func MapToEnvironment(envMap map[string]string) []string {
	var envStrings []string
	for k, v := range envMap {
		envStrings = append(envStrings, k+"="+v)
	}
	return envStrings
}
