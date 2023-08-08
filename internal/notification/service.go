// service.go
package notification

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/troke12/docomon/internal/container"
	"github.com/docker/go-connections/nat"
)

// NotificationService implements the NotificationService interface.
type NotificationServiceImpl struct {
	containerService container.ContainerService
	webhookService   WebhookService
}

// NewNotificationService creates a new NotificationService instance.
func NewNotificationService(containerService container.ContainerService, webhookService WebhookService) NotificationService {
	return &NotificationServiceImpl{
		containerService: containerService,
		webhookService:   webhookService,
	}
}

// CompareContainersAndNotify compares containers and sends notifications.
func (ns *NotificationServiceImpl) CompareContainersAndNotify(initialContainers, currentContainers []types.Container) error {
	initialContainerMap := make(map[string]struct{})
	currentContainerMap := make(map[string]struct{})

	for _, c := range initialContainers {
		initialContainerMap[c.ID] = struct{}{}
	}

	for _, c := range currentContainers {
		currentContainerMap[c.ID] = struct{}{}
	}

	// Check for new and removed containers
	for _, c := range currentContainers {
		if _, exists := initialContainerMap[c.ID]; !exists {
			// New container detected, send notifications to both Discord and Google Chat
			containerInfo, err := ns.containerService.InspectContainer(context.Background(), c.ID)
			if err != nil {
				log.Println("Error inspecting container:", err)
				continue
			}

			portMappings := getContainerPorts(c.ID, containerInfo.HostConfig.PortBindings)
			message := formatContainerMessage(c.ID[:12], c.Names[0], c.Image, portMappings)
			ns.sendWebhooks("New container started:\n" + message)
		}
	}

	for _, c := range initialContainers {
		if _, exists := currentContainerMap[c.ID]; !exists {
			// Container removed, send notifications to both Discord and Google Chat
			containerInfo, err := ns.containerService.InspectContainer(context.Background(), c.ID)
			if err != nil {
				log.Println("Error inspecting container:", err)
				continue
			}

			portMappings := getContainerPorts(c.ID, containerInfo.HostConfig.PortBindings)
			message := formatContainerMessage(c.ID[:12], c.Names[0], c.Image, portMappings)
			ns.sendWebhooks("Removed container:\n" + message)
		}
	}

	return nil
}

func (ns *NotificationServiceImpl) sendWebhooks(message string) {
	go func() {
		if ns.webhookService.GetDiscordWebhookURL() != "" {
			err := ns.webhookService.SendWebhook(message)
			if err != nil {
				log.Println("Error sending Discord webhook:", err)
			}
		}

		if ns.webhookService.GetGoogleChatWebhookURL() != "" {
			err := ns.webhookService.SendWebhook(message)
			if err != nil {
				log.Println("Error sending Google Chat webhook:", err)
			}
		}
	}()
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
	hostname, _ := os.Hostname() // Retrieve the hostname of the machine

	message := fmt.Sprintf("ID: %s\nName: %s\nImage: %s\nPorts: %s\nHost: %s\n", id, escapeSpecialChars(name), image, strings.Join(portMappings, ", "), hostname)
	return message
}
