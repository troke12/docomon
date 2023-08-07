package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	discordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
    googleChatWebhookURL = os.Getenv("GOOGLE_CHAT_WEBHOOK_URL")
)

func main() {
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

	for {
		fmt.Printf("Time: %s\n", time.Now().Format("2 January 2006 3:04 PM"))

		currentContainers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			log.Println("Error listing containers:", err)
			continue
		}

		displayContainers(cli, currentContainers)
		compareContainersAndNotify(cli, initialContainers, currentContainers)

		initialContainers = currentContainers

		time.Sleep(5 * time.Second)
		clearTerminal()
		fmt.Println("Container monitoring:")
	}
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func displayContainerPorts(cli *client.Client, containerID string) {
	ctx := context.Background()

	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Println("Error inspecting container:", err)
		return
	}

	displayPorts(inspect.NetworkSettings.Ports)
}

func displayContainers(cli *client.Client, containers []types.Container) {
	fmt.Println("Current containers:")
	for _, c := range containers {
		fmt.Printf("ID: %s, Name: %s, Image: %s, ", c.ID[:12], c.Names[0], c.Image)
		displayContainerPorts(cli, c.ID)
		fmt.Println()
	}
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

func displayPorts(portBindings map[nat.Port][]nat.PortBinding) {
	for p, bindings := range portBindings {
		for _, binding := range bindings {
			portMapping := fmt.Sprintf("%s->%s", binding.HostPort, p.Port())
			fmt.Printf("Port: %s\n", portMapping)
		}
	}
}

func compareContainersAndNotify(cli *client.Client, initialContainers, currentContainers []types.Container) {
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
            containerInfo, err := cli.ContainerInspect(context.Background(), c.ID)
            if err != nil {
                log.Println("Error inspecting container:", err)
                continue
            }

            portMappings := getContainerPorts(c.ID, containerInfo.HostConfig.PortBindings)
            message := formatContainerMessage(c.ID[:12], c.Names[0], c.Image, portMappings)
            sendDiscordWebhook("New container started:\n" + message)
            sendGoogleChatWebhook("New container started:\n" + message)
        }
    }

    for _, c := range initialContainers {
        if _, exists := currentContainerMap[c.ID]; !exists {
            // Container removed, send notifications to both Discord and Google Chat
            containerInfo, err := cli.ContainerInspect(context.Background(), c.ID)
            if err != nil {
                log.Println("Error inspecting container:", err)
                continue
            }

            portMappings := getContainerPorts(c.ID, containerInfo.HostConfig.PortBindings)
            message := formatContainerMessage(c.ID[:12], c.Names[0], c.Image, portMappings)
            sendDiscordWebhook("Removed container:\n" + message)
            sendGoogleChatWebhook("Removed container:\n" + message)
        }
    }
}



func formatContainerMessage(id string, name string, image string, portMappings []string) string {
    hostname, _ := os.Hostname() // Retrieve the hostname of the machine
    
    message := fmt.Sprintf("ID: %s\nName: %s\nImage: %s\nPorts: %s\nHost: %s\n", id, escapeSpecialChars(name), image, strings.Join(portMappings, ", "), hostname)
    return message
}

func sendDiscordWebhook(message string) {
	// Remove newline characters from the message
	message = strings.ReplaceAll(message, "\n", " ")

	// Escape special characters
	formattedMessage := escapeSpecialChars(message)
	payload := []byte(fmt.Sprintf(`{"content": "%s"}`, formattedMessage))

	//fmt.Println("Sending payload:", string(payload)) // Debug: Print the payload being sent

	resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error sending Discord webhook:", err)
		return
	}
	defer resp.Body.Close()

	//fmt.Println("Webhook response status:", resp.Status) // Debug: Print the response status

	if resp.StatusCode != http.StatusNoContent {
		log.Printf("Unexpected response from Discord webhook: %s", resp.Status)
	}
}

func sendGoogleChatWebhook(message string) {
	// Remove newline characters from the message
	message = strings.ReplaceAll(message, "\n", " ")

	// Escape special characters
	formattedMessage := escapeSpecialChars(message)
	payload := []byte(fmt.Sprintf(`{"text": "%s"}`, formattedMessage))

	resp, err := http.Post(googleChatWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Error sending Google Chat webhook:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected response from Google Chat webhook: %s", resp.Status)
	}
}

func escapeSpecialChars(input string) string {
	// Escape special characters for JSON payload
	input = strings.ReplaceAll(input, `\`, `\\`)
	input = strings.ReplaceAll(input, `"`, `\"`)
	return input
}