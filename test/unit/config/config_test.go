package config_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	"subscriptions/src/config"
	"testing"
)

func TestLoadingProfile(t *testing.T) {
	config.LoadProfileFromFile("./test-profiles/test.json", "test")
	activeConfig := config.GetConfig()

	assert.Equal(t, "test", config.GetProfileName())

	assert.Equal(t, 1234, activeConfig.Server.Port)
	assert.Equal(t, "host", activeConfig.Database.Host)
	assert.Equal(t, true, activeConfig.Logging.DevelopmentLogger)
}

func TestLoadingProfileWithEnvironmentVariableOverride(t *testing.T) {
	os.Setenv("DATABASE_USER", "override-user")

	config.LoadProfileFromFile("./test-profiles/test.json", "test")
	activeConfig := config.GetConfig()

	assert.Equal(t, "override-user", activeConfig.Database.User)
}
