package docker

import (
	"os"
	"path/filepath"
	"strings"
)

// DockerInfo contains information about Docker files in a project
type DockerInfo struct {
	HasDockerfile bool
	HasCompose    bool
	Dockerfiles   []string // All Dockerfile variants found
	ComposeFiles  []string // All compose files found
}

// composeFileNames are exact filenames for Docker Compose files
var composeFileNames = []string{
	"docker-compose.yml",
	"docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
}

// Detect scans a project directory for Docker-related files
func Detect(projectPath string) (*DockerInfo, error) {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil, err
	}

	info := &DockerInfo{
		Dockerfiles:  make([]string, 0),
		ComposeFiles: make([]string, 0),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Check for Dockerfile
		if isDockerfile(name) {
			info.HasDockerfile = true
			info.Dockerfiles = append(info.Dockerfiles, name)
		}

		// Check for Compose file
		if isComposeFile(name) {
			info.HasCompose = true
			info.ComposeFiles = append(info.ComposeFiles, name)
		}
	}

	return info, nil
}

// isDockerfile checks if a filename is a Dockerfile
func isDockerfile(filename string) bool {
	// Exact match
	if filename == "Dockerfile" {
		return true
	}

	// Dockerfile.* pattern
	if strings.HasPrefix(filename, "Dockerfile.") {
		return true
	}

	return false
}

// isComposeFile checks if a filename is a Docker Compose file
func isComposeFile(filename string) bool {
	for _, composeName := range composeFileNames {
		if filename == composeName {
			return true
		}
	}
	return false
}

// GetComposeCommand returns the preferred compose command
// Checks for docker compose v2 plugin first, falls back to docker-compose v1
func GetComposeCommand() string {
	// Modern Docker CLI with compose plugin
	return "docker compose"
}

// GetPrimaryDockerfile returns the most likely primary Dockerfile
func GetPrimaryDockerfile(dockerfiles []string) string {
	// Prefer plain "Dockerfile" over variants
	for _, df := range dockerfiles {
		if df == "Dockerfile" {
			return df
		}
	}

	// Otherwise return first one found
	if len(dockerfiles) > 0 {
		return dockerfiles[0]
	}

	return "Dockerfile"
}

// GetPrimaryComposeFile returns the most likely primary compose file
func GetPrimaryComposeFile(composeFiles []string) string {
	// Preference order: compose.yml > docker-compose.yml > compose.yaml > docker-compose.yaml
	preferences := []string{"compose.yml", "docker-compose.yml", "compose.yaml", "docker-compose.yaml"}

	for _, pref := range preferences {
		for _, cf := range composeFiles {
			if cf == pref {
				return cf
			}
		}
	}

	// Otherwise return first one found
	if len(composeFiles) > 0 {
		return composeFiles[0]
	}

	return "docker-compose.yml"
}

// HasDevContainer checks if a project has a dev container configuration
func HasDevContainer(projectPath string) bool {
	devcontainerPath := filepath.Join(projectPath, ".devcontainer", "devcontainer.json")
	_, err := os.Stat(devcontainerPath)
	return err == nil
}
