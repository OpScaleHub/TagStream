package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

const (
	githubRepo  = "OpScaleHub/TagStream"
	composeFile = "docker-compose.yml"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func main() {
	latestTag, err := getLatestReleaseTag()
	if err != nil {
		fmt.Println("Error fetching latest release:", err)
		return
	}

	fmt.Println("Latest release tag:", latestTag)

	err = updateDockerCompose()
	if err != nil {
		fmt.Println("Error updating Docker Compose:", err)
		return
	}

	fmt.Println("Docker Compose updated successfully.")
}

func getLatestReleaseTag() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var release Release
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", err
	}

	return release.TagName, nil
}

func updateDockerCompose() error {
	cmd := exec.Command("docker", "compose", "-f", composeFile, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("docker", "compose", "-f", composeFile, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
