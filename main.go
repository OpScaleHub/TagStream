package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/go-github/v45/github"
)

var (
	latestTag string
	mu        sync.Mutex
)

func main() {
	ctx := context.Background()
	client := github.NewClient(nil)

	// Replace with your GitHub repository owner and name
	owner := "OpScaleHub"
	repo := "TagStream"

	// Check the latest release tag initially
	checkLatestRelease(ctx, client, owner, repo)

	// Start a parallel thread to subscribe to new release tags
	go subscribeToReleases(ctx, client, owner, repo)

	// Keep the main thread running
	select {}
}

func checkLatestRelease(ctx context.Context, client *github.Client, owner, repo string) {
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		log.Fatalf("Error fetching latest release: %v", err)
	}

	mu.Lock()
	latestTag = release.GetTagName()
	mu.Unlock()

	fmt.Printf("Latest release tag: %s\n", latestTag)
}

func subscribeToReleases(ctx context.Context, client *github.Client, owner, repo string) {
	for {
		time.Sleep(1 * time.Minute) // Check for new releases every minute

		release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			log.Printf("Error fetching latest release: %v", err)
			continue
		}

		mu.Lock()
		if release.GetTagName() != latestTag {
			latestTag = release.GetTagName()
			mu.Unlock()

			fmt.Printf("New release tag detected: %s\n", latestTag)
			handleNewRelease(ctx, owner, repo)
		} else {
			mu.Unlock()
		}
	}
}

func handleNewRelease(ctx context.Context, owner, repo string) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/docker-compose.yml", owner, repo, latestTag)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching docker-compose.yml: %v", err)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading docker-compose.yml: %v", err)
		return
	}

	err = os.WriteFile("docker-compose.yml", data, 0644)
	if err != nil {
		log.Printf("Error writing docker-compose.yml: %v", err)
		return
	}

	cmd := exec.Command("docker", "compose", "down")
	err = cmd.Run()
	if err != nil {
		log.Printf("Error stopping Docker Compose: %v", err)
		return
	}

	cmd = exec.Command("docker", "compose", "up", "-d")
	err = cmd.Run()
	if err != nil {
		log.Printf("Error starting Docker Compose: %v", err)
		return
	}

	fmt.Println("Docker Compose restarted successfully.")
}
