package notification

import (
	"bytes"
	//"fmt"
	"net/http"
	"strings"
)

type WebhookClient struct {
	googleChatURL string
	discordURL    string
}

func NewWebhookClient(googleChatURL, discordURL string) *WebhookClient {
	return &WebhookClient{
		googleChatURL: googleChatURL,
		discordURL:    discordURL,
	}
}

func (wc *WebhookClient) SendWebhook(webhookURL, payload string) error {
	client := http.DefaultClient

	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()


	// this is needed when debugging the webhook
	// if resp.StatusCode != http.StatusNoContent {
	// 	return fmt.Errorf("unexpected response from webhook: %s", resp.Status)
	// }

	return nil
}


func (wc *WebhookClient) GetDiscordWebhookURL() string {
	return wc.discordURL
}

func (wc *WebhookClient) GetGoogleChatWebhookURL() string {
	return wc.googleChatURL
}

func escapeSpecialChars(input string) string {
	input = strings.ReplaceAll(input, `\`, `\\`)
	input = strings.ReplaceAll(input, `"`, `\"`)
	return input
}
