package security

import (
	"encoding/json"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2"
	"strings"
	"subscriptions/src/config"
	db "subscriptions/src/database"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
	"time"
)

var apiKeysCache []models.ApiKeyPermission
var apiKeysCacheExpiry time.Time

func CheckApiKey(monitoringContext *monitoring.Context, key string, permission string) (keyMatched bool, permissionMatched bool, clientName string) {
	if apiKeysCacheExpiry.Before(time.Now()) {
		var err error
		apiKeysCache, err = db.GetApiKeys(monitoringContext)
		if err != nil {
			monitoringContext.Error("Unable to get API Key Permissions", zap.Error(err))
			return false, false, ""
		}

		apiKeysCacheExpiry = time.Now().Add(time.Duration(config.GetConfig().AuthConfig.ApiKeyCacheMs) * time.Millisecond)
	}

	for _, keyPermission := range apiKeysCache {
		if keyPermission.ApiKey == key {
			keyMatched = true
		}

		if keyPermission.ApiKey == key && keyPermission.Permission != nil && *keyPermission.Permission == permission {
			return true, true, keyPermission.Owner
		}
	}

	return keyMatched, false, ""
}

type OAuth2ServiceJwt struct {
	//TODO: Awaiting actual shape of this from Enrique
	Sub            string
	Name           string
	SubscriptionId string `json:"subscription_id"`
	AccountId      string `json:"account_id"`
	AndroidId      string `json:"android_id"`
}

func CheckJwt(monitoringContext *monitoring.Context, encodedJwt string) (passed bool, subscriptionId, accountId, androidId string) {
	if count := strings.Count(encodedJwt, "."); count != 2 {
		return false, "", "", ""
	}

	jwt, err := jose.ParseSigned(encodedJwt)
	if err != nil {
		monitoringContext.Error("Could not parse JWT", zap.Error(err))
		return false, "", "", ""
	}

	//Caution: This assumes that the authentication middleware on Krakend has verified the JWT's signature against the
	//		   OAuth2 server's public key already.

	var result OAuth2ServiceJwt
	err = json.Unmarshal(jwt.UnsafePayloadWithoutVerification(), &result)
	if err != nil {
		monitoringContext.Error("Unable to deserialize JWT payload: %w", zap.Error(err))
		return false, "", "", ""
	}

	return true, result.SubscriptionId, result.AccountId, result.AndroidId
}
