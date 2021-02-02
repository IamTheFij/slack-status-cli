package main

import "os"

func getEnvOrDefault(name, defaultValue string) string {
	val, ok := os.LookupEnv(name)
	if ok {
		return val
	}

	return defaultValue
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}
