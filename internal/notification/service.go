// service.go
package notification

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/troke12/docomon/internal/container"
)

type NotificationServiceImpl struct {
	containerService container.ContainerService
	webhookService   WebhookService
}

func NewNotificationService(containerService container.ContainerService, webhookService WebhookService) NotificationService {
	return &NotificationServiceImpl{
		containerService: containerService,
		webhookService:   webhookService,
	}
}

func (ns *NotificationServiceImpl) CompareContainersAndNotify(initialContainers, currentContainers []types.Container) error {
	initialContainerMap := make(map[string]struct{})
	currentContainerMap := make(map[string]struct{})
	processedContainers := make(map[string]bool)

	for _, c := range initialContainers {
		initialContainerMap[c.ID] = struct{}{}
	}

	for _, c := range currentContainers {
		currentContainerMap[c.ID] = struct{}{}
	}

	for _, c := range currentContainers {
		if processedContainers[c.ID] {
			// Skip containers that have already been processed
			continue
		}

		containerInfo, err := ns.containerService.InspectContainer(context.Background(), c.ID)
		if err != nil {
			log.Println("Error inspecting container:", err)
			continue
		}

		portMappings := getContainerPorts(c.ID, containerInfo.HostConfig.PortBindings)
		message := formatContainerMessage(c.ID[:12], c.Names[0], c.Image, portMappings)

		if _, exists := initialContainerMap[c.ID]; exists {
			currentContainerMap[c.ID] = struct{}{} // Mark container as present
		} else {
			// New container detected
			message = "New container started: " + message

			discordWebhookURL := ns.webhookService.GetDiscordWebhookURL()
			if discordWebhookURL != "" {
				if err := ns.sendDiscordWebhook(message, discordWebhookURL); err != nil {
					log.Println("Error sending Discord webhook:", err)
				}
			}

			googleChatWebhookURL := ns.webhookService.GetGoogleChatWebhookURL()
			if googleChatWebhookURL != "" {
				if err := ns.sendGoogleChatWebhook(message, googleChatWebhookURL); err != nil {
					log.Println("Error sending Google Chat webhook:", err)
				}
			}
		}

		// Mark container as processed
		processedContainers[c.ID] = true
	}

	for _, v := range initialContainers {
		if _, exists := currentContainerMap[v.ID]; !exists {
			// Container removed
			containerInfo, err := ns.containerService.InspectContainer(context.Background(), v.ID)
			if err != nil {
				log.Println("Error inspecting container:", err)
				continue
			}

			portMappings := getContainerPorts(v.ID, containerInfo.HostConfig.PortBindings)
			message := formatContainerMessage(v.ID[:12], v.Names[0], v.Image, portMappings)
			message = "Removed container: " + message

			discordWebhookURL := ns.webhookService.GetDiscordWebhookURL()
			if discordWebhookURL != "" {
				if err := ns.sendDiscordWebhook(message, discordWebhookURL); err != nil {
					log.Println("Error sending Discord webhook:", err)
				}
			}

			googleChatWebhookURL := ns.webhookService.GetGoogleChatWebhookURL()
			if googleChatWebhookURL != "" {
				if err := ns.sendGoogleChatWebhook(message, googleChatWebhookURL); err != nil {
					log.Println("Error sending Google Chat webhook:", err)
				}
			}
		}
	}

	return nil
}


func (ns *NotificationServiceImpl) sendDiscordWebhook(message, webhookURL string) error {
	//log.Printf("Sending Discord webhook:\nURL: %s\nMessage: %s\n", webhookURL, createDiscordPayload(message))
	return ns.webhookService.SendWebhook(webhookURL, createDiscordPayload(message))
}

func (ns *NotificationServiceImpl) sendGoogleChatWebhook(message, webhookURL string) error {
	//log.Printf("Sending Google Chat webhook:\nURL: %s\nMessage: %s\n", webhookURL, createGooglePayload(message)) // Debug log webhook
	return ns.webhookService.SendWebhook(webhookURL, createGooglePayload(message))
}

func getContainerPorts(containerID string, portBindings map[nat.Port][]nat.PortBinding) []string {
	var portMappings []string

	for p, bindings := range portBindings {
		for _, binding := range bindings {
			portMapping := fmt.Sprintf("%s->%s", binding.HostPort, p.Port())
			portMappings = append(portMappings, portMapping)
		}
	}

	return portMappings
}

func formatContainerMessage(id string, name string, image string, portMappings []string) string {
	serverName := os.Getenv("SERVER_NAME")
	if serverName == "" {
		log.Println("SERVER_NAME environment variable not set")
	}

	message := fmt.Sprintf("ID: %s, Name: %s, Image: %s, Ports: %s, Server: %s", id, escapeSpecialChars(name), image, strings.Join(portMappings, ", "), serverName)
	return message
}

func createDiscordPayload(message string) string {
	return fmt.Sprintf(`{"content": "%s"}`, escapeSpecialChars(message))
}

func createGooglePayload(message string) string {
	return fmt.Sprintf(`{"text": "%s"}`, escapeSpecialChars(message))
}
