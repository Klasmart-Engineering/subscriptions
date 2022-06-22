package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"subscriptions/src/utils"
)

var activeConfig *config
var activeProfile *string

type config struct {
	Server         serverConfig
	Logging        loggingConfig
	Database       databaseConfig
	NewRelicConfig newRelicConfig
	AuthConfig     authConfig
	AwsConfig      awsConfig
	BucketConfig   bucketConfig
	Testing        bool
}

type serverConfig struct {
	Port int
}

type loggingConfig struct {
	DevelopmentLogger bool
}

type newRelicConfig struct {
	EntityName    string
	Enabled       bool
	LicenseKey    string
	TracerEnabled bool
}

type databaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	Seed         bool
}

type authConfig struct {
	ApiKeyCacheMs int
}

type awsConfig struct {
	Region          string
	ManuallySpecify bool
	AccessKeyId     *string
	AccessKeySecret *string
	Endpoint        *string
}

type bucketConfig struct {
	AccessLogBucket string
}

func LoadProfile(name string) {
	LoadProfileFromFile(fmt.Sprintf("./profiles/%s.json", name), name)
}

func LoadProfileFromFile(file string, name string) {
	var contents = readFile(file)

	err := json.NewDecoder(strings.NewReader(contents)).Decode(&activeConfig)
	if err != nil {
		panic(fmt.Sprintf("Could not deserialise config file: %s", err))
	}

	activeProfile = &name

	replaceFromEnvironmentVariables("", activeConfig)
}

func replaceFromEnvironmentVariables(path string, thing interface{}) {
	configValue := reflect.Indirect(reflect.ValueOf(thing))
	configType := reflect.Indirect(configValue).Type()

	i := 0
	for i < configType.NumField() {
		fieldName := strings.ToUpper(configType.Field(i).Name)
		if path != "" {
			fieldName = path + "_" + fieldName
		}

		fieldType := configType.Field(i).Type

		if fieldType.Kind() == reflect.Struct {
			replaceFromEnvironmentVariables(fieldName, configValue.Field(i).Addr().Interface())
		}

		if fieldType.Kind() == reflect.String {
			value := utils.GetStringEnv(fieldName)
			if value != nil {
				configValue.Field(i).SetString(*value)
			}
		}

		if fieldType.Kind() == reflect.Int {
			value := utils.GetIntEnv(fieldName)
			if value != nil {
				configValue.Field(i).SetInt(int64(*value))
			}
		}

		if fieldType.Kind() == reflect.Bool {
			value := utils.GetBoolEnv(fieldName)
			if value != nil {
				configValue.Field(i).SetBool(*value)
			}
		}

		i++
	}
}

func GetConfig() *config {
	if activeConfig == nil {
		log.Panicf("Attempt to get config before any profile loaded")
	}

	return activeConfig
}

func GetProfileName() string {
	if activeProfile == nil {
		log.Panicf("Attempt to get profile name before any profile loaded")
	}

	return *activeProfile
}

func readFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		log.Panicf("Could not read file %s", file)
	}
	return string(content)
}
