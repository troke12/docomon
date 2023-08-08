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
)

func main() {
	discordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	googleChatWebhookURL := os.Getenv("GOOGLE_CHAT_WEBHOOK_URL")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal("Error creating Docker client:", err)
	}

	ctx := context.Background()

	initialContainers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Fatal("Error listing initial containers:", err)
	}

	clearTerminal()
	fmt.Println("Container monitoring:")

	containerService := container.NewContainerService(cli)
	webhookClient := notification.NewWebhookClient(googleChatWebhookURL, discordWebhookURL)
	notificationService := notification.NewNotificationService(containerService, webhookClient)

	for {
		fmt.Printf("Time: %s\n", time.Now().Format("2 January 2006 3:04 PM"))

		currentContainers, err := containerService.ListContainers(ctx, types.ContainerListOptions{})
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

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
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

func displayContainers(containerService container.ContainerService, containers []types.Container) {
	fmt.Println("Current containers:")
	for _, c := range containers {
		fmt.Printf("ID: %s, Name: %s, Image: %s, ", c.ID[:12], c.Names[0], c.Image)
		displayContainerPorts(containerService, c.ID)
		fmt.Println()
	}
}

func displayPorts(portBindings map[nat.Port][]nat.PortBinding) {
	for p, bindings := range portBindings {
		for _, binding := range bindings {
			portMapping := fmt.Sprintf("%s->%s", binding.HostPort, p.Port())
			fmt.Printf("Port: %s\n", portMapping)
		}
	}
}
