# Docker Container Monitoring

This repository contains a container monitoring application that provides real-time monitoring and notifications for Docker containers. The application tracks the state of containers, detects new containers, and sends notifications when containers are added or removed.

Features:
- Real-time monitoring of Docker containers
- Detection of new containers and notification alerts
- Integration with Discord and Google Chat for sending notifications
- Detailed container information including ID, name, image, and port bindings

The application is built using Go and leverages the Docker API for container management. It utilizes webhooks to send notifications to Discord and Google Chat platforms. The codebase will be rewrite in the future to clean architecture design pattern, allowing for easy testing, maintainability, and future enhancements.

Get started with container monitoring and receive timely notifications for your Docker environment using this Docker Container Monitoring application.

## Environment
```
SERVER_NAME=troke-server
GOOGLE_CHAT_WEBHOOK_URL=your_google_chat_webhook_url
DISCORD_WEBHOOK_URL=your_discord_webhook_url
```

Define the environment just what you want

## Using docker

- Run this command
```
docker run --restart always -v /var/run/docker.sock:/var/run/docker.sock --env-file .env ---name docomon troke12/docomon:latest
```
- Make sure you have .env file created and point the path on `--env-file /path/to/your/env`
- Make sure you have to change the `/var/run/docker.sock` to the right path of docker sock

## Manual setup

This setup was currently on linux

- Download the specific file on [release](https://github.com/troke12/docomon/releases/latest)
- Copy the file `cp docomon-linux-amd64 /user/local/bin/docomon`
- Set the permission `chmod +x /usr/local/bin/docomon`
- Create `docomon.conf` on `/etc/docomon.conf`
```
cat <<EOF > "/etc/docomon.conf"
SERVER_NAME="server-name"
GOOGLE_CHAT_WEBHOOK_URL="https://chat.googleapis.com"
DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/"
EOF
```
- Create systemd service named `docomon.service`
```
cat <<EOF > "/etc/systemd/system/docomon.service"
[Unit]
Description=docomon Service
After=network.target

[Service]
ExecStart=/usr/local/bin/docomon --env-file=/etc/docomon.conf

[Install]
WantedBy=multi-user.target
EOF
```
- systemctl daemon-reload
- systemctl enable docomon.service
- systemctl start docomon.service

## Screenshot

### Discord

![image](https://cdn.discordapp.com/attachments/1011830399032369212/1138115372193685575/image.png)

### Google Chat

![img](https://cdn.discordapp.com/attachments/1011830399032369212/1138115512757403668/image.png)

### Terminal

![terminal](https://cdn.discordapp.com/attachments/1011830399032369212/1138116185582485504/image.png)