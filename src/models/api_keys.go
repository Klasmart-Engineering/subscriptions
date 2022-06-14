package models

type ApiKeyPermission struct {
	Owner      string
	ApiKey     string
	Permission *string
}
