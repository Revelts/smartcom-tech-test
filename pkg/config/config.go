package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func GetEnv(key, defaultValue string) (value string) {
	value = os.Getenv(key)
	if value != "" {
		return
	}
	value = defaultValue
	return
}

func GetEnvInt(key string, defaultValue int) (result int) {
	value := os.Getenv(key)
	if value != "" {
		var err error
		result, err = strconv.Atoi(value)
		if err == nil {
			return
		}
	}
	result = defaultValue
	return
}

func GetEnvBool(key string, defaultValue bool) (result bool) {
	value := os.Getenv(key)
	if value != "" {
		var err error
		result, err = strconv.ParseBool(value)
		if err == nil {
			return
		}
	}
	result = defaultValue
	return
}

func GetEnvDuration(key string, defaultValue time.Duration) (result time.Duration) {
	value := os.Getenv(key)
	if value != "" {
		var err error
		result, err = time.ParseDuration(value)
		if err == nil {
			return
		}
	}
	result = defaultValue
	return
}

func MustGetEnv(key string) (value string) {
	value = os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return
}
