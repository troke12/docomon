package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/troke12/docomon/internal/container"
	"github.com/troke12/docomon/internal/notification"
	"github.com/joho/godotenv"
)

func main() {
	// Add an .env
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file:", err)
    }

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal("Error creating Docker client:", err)
	}

	containerService := container.NewContainerService(dockerClient) // Create the container service
	webhookClient := setupWebhookClient()

	notificationService := notification.NewNotificationService(containerService, webhookClient)

	ctx := context.Background()

	initialContainers, err := containerService.ListContainers(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Fatal("Error listing initial containers:", err)
	}

	clearTerminal()
	fmt.Println("Container monitoring:")

	for {
		displayTimestamp()
		currentContainers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			log.Println("Error listing containers:", err)
			continue
		}

		displayContainers(containerService, currentContainers)
		notificationService.CompareContainersAndNotify(initialContainers, currentContainers)

		initialContainers = currentContainers

		time.Sleep(5 * time.Second)
		clearTerminal()
		fmt.Println("Container monitoring:")
	}
}

func setupWebhookClient() notification.WebhookService {
	googleChatURL := os.Getenv("GOOGLE_CHAT_WEBHOOK_URL")
	discordURL := os.Getenv("DISCORD_WEBHOOK_URL")
	return notification.NewWebhookClient(googleChatURL, discordURL)
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func displayTimestamp() {
	fmt.Printf("Time: %s\n", time.Now().Format("2 January 2006 3:04 PM"))
}

func displayContainers(containerService container.ContainerService, containers []types.Container) {
	fmt.Println("Current containers:")
	for _, c := range containers {
		displayContainerInfo(containerService, c)
		fmt.Println()
	}
}

func displayContainerInfo(containerService container.ContainerService, c types.Container) {
	fmt.Printf("ID: %s, Name: %s, Image: %s, ", c.ID[:12], c.Names[0], c.Image)
	displayContainerPorts(containerService, c.ID)
}

func displayContainerPorts(containerService container.ContainerService, containerID string) {
	ctx := context.Background()
	inspect, err := containerService.InspectContainer(ctx, containerID)
	if err != nil {
		log.Println("Error inspecting container:", err)
		return
	}

	displayPorts(inspect.NetworkSettings.Ports)
}

func displayPorts(portBindings map[nat.Port][]nat.PortBinding) {
	for p, bindings := range portBindings {
		for _, binding := range bindings {
			portMapping := fmt.Sprintf("%s->%s", binding.HostPort, p.Port())
			fmt.Printf("Port: %s\n", portMapping)
		}
	}
}
