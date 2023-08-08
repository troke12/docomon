// notification.go
package notification

import (
	"github.com/docker/docker/api/types"
)

// WebhookService defines the methods for sending webhooks.
type WebhookService interface {
	SendWebhook(webhookURL, message string) error
	GetDiscordWebhookURL() string
	GetGoogleChatWebhookURL() string
}

// NotificationService defines the methods for handling notifications.
type NotificationService interface {
	CompareContainersAndNotify(initialContainers, currentContainers []types.Container) error
}

