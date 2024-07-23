#!/bin/sh

# Variables
GITHUB_REPO="username/repository"
COMPOSE_FILE="docker-compose.yml"

# Get the latest release tag from GitHub
LATEST_TAG=$(curl --silent "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | jq -r .tag_name)

# Pull the latest images
docker-compose -f $COMPOSE_FILE pull

# Restart the services
docker-compose -f $COMPOSE_FILE up -d
