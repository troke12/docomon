// webhook_client.go
package notification

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

// WebhookClient is responsible for sending webhooks.
type WebhookClient struct {
	googleChatURL string
	discordURL    string
}

// NewWebhookClient creates a new WebhookClient instance.
func NewWebhookClient(googleChatURL, discordURL string) *WebhookClient {
	return &WebhookClient{
		googleChatURL: googleChatURL,
		discordURL:    discordURL,
	}
}

// SendWebhook sends the webhook message.
func (wc *WebhookClient) SendWebhook(message string) error {
	// Remove newline characters from the message
	message = strings.ReplaceAll(message, "\n", " ")

	// Escape special characters
	formattedMessage := escapeSpecialChars(message)
	payload := []byte(fmt.Sprintf(`{"content": "%s"}`, formattedMessage))

	client := http.DefaultClient

	resp, err := client.Post(wc.discordURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected response from webhook: %s", resp.Status)
	}

	return nil
}

func escapeSpecialChars(input string) string {
	// Escape special characters for JSON payload
	input = strings.ReplaceAll(input, `\`, `\\`)
	input = strings.ReplaceAll(input, `"`, `\"`)
	return input
}
