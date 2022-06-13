package security

import (
	"go.uber.org/zap"
	"subscriptions/src/config"
	db "subscriptions/src/database"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
	"time"
)

var apiKeysCache []models.ApiKeyPermission
var apiKeysCacheExpiry time.Time

func CheckApiKey(monitoringContext *monitoring.Context, key string, permission string) (keyMatched bool, permissionMatched bool) {
	if apiKeysCacheExpiry.Before(time.Now()) {
		var err error
		apiKeysCache, err = db.GetApiKeys(monitoringContext)
		if err != nil {
			monitoringContext.Error("Unable to get API Key Permissions", zap.Error(err))
			return false, false
		}

		apiKeysCacheExpiry = time.Now().Add(time.Duration(config.GetConfig().AuthConfig.ApiKeyCacheMs) * time.Millisecond)
	}

	for _, keyPermission := range apiKeysCache {
		if keyPermission.ApiKey == key {
			keyMatched = true
		}

		if keyPermission.ApiKey == key && keyPermission.Permission == permission {
			return true, true
		}
	}

	return keyMatched, false
}
