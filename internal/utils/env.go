package utils

import "os"

// GetEnv returns the value of the environment variable named by key.
// If the variable is not set or is empty, def is returned.
func GetEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// MustGetEnv returns the value of the environment variable named by key.
// Panics if the variable is not set — use this for required secrets at startup.
func MustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}
