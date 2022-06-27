package utils

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

func GetStringEnv(key string) *string {
	value, exists := os.LookupEnv(key)

	if !exists {
		return nil
	}

	return &value
}

func GetIntEnv(key string) *int {
	value, exists := os.LookupEnv(key)

	if !exists {
		return nil
	}

	parsedValue, err := strconv.Atoi(value)

	if err != nil {
		log.Panicf("Could not parse int for environment variable %s", key)
	}

	return &parsedValue
}

func GetBoolEnv(key string) *bool {
	value, exists := os.LookupEnv(key)

	if !exists {
		return nil
	}

	parsedValue, err := strconv.ParseBool(value)

	if err != nil {
		log.Panicf("Could not parse bool for environment variable %s", key)
	}

	return &parsedValue
}

func MustGetEnvOrFlag(key string) string {
	if val := os.Getenv(strings.ToUpper(key)); "" != val {
		return val
	}

	flagValue := flag.String(key, "", "")
	flag.Parse()

	args := flag.Args()

	if *flagValue == "" {
		log.Panicf("Could not get env or flag for %s, %s", key, args)
	}

	return *flagValue
}
