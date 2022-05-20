package instrument

import (
	"fmt"
	"os"
	"strings"
)

func MustGetEnv(key string) string {
	if val := os.Getenv(key); "" != val {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}

func GetBrokers() []string {
	return strings.Split(MustGetEnv("BROKERS"), ",")
}
