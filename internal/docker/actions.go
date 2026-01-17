package docker

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Action represents a Docker action that can be executed
type Action struct {
	ID          string
	Name        string
	Description string
	RequireFile bool // Requires Dockerfile
	RequireComp bool // Requires Compose file
}

// AvailableActions returns all available Docker actions
var AvailableActions = []Action{
	{
		ID:          "docker-build",
		Name:        "🏗️  Build Image",
		Description: "Build Docker image from Dockerfile",
		RequireFile: true,
	},
	{
		ID:          "docker-run",
		Name:        "▶️  Run Container",
		Description: "Run container interactively",
		RequireFile: true,
	},
	{
		ID:          "docker-run-detached",
		Name:        "🔄 Run Detached",
		Description: "Run container in background",
		RequireFile: true,
	},
	{
		ID:          "docker-ps",
		Name:        "📋 List Containers",
		Description: "Show running containers",
		RequireFile: false,
	},
	{
		ID:          "docker-logs",
		Name:        "📜 View Logs",
		Description: "Stream container logs",
		RequireFile: false,
	},
	{
		ID:          "compose-up",
		Name:        "🚀 Compose Up",
		Description: "Start all services",
		RequireComp: true,
	},
	{
		ID:          "compose-up-detached",
		Name:        "🔄 Compose Up (Detached)",
		Description: "Start services in background",
		RequireComp: true,
	},
	{
		ID:          "compose-down",
		Name:        "🛑 Compose Down",
		Description: "Stop and remove services",
		RequireComp: true,
	},
	{
		ID:          "compose-build",
		Name:        "🏗️  Compose Build",
		Description: "Build all images",
		RequireComp: true,
	},
	{
		ID:          "compose-logs",
		Name:        "📜 Compose Logs",
		Description: "Stream all service logs",
		RequireComp: true,
	},
	{
		ID:          "compose-ps",
		Name:        "📋 Compose PS",
		Description: "List services",
		RequireComp: true,
	},
}

// ExecuteResult represents the result of executing a Docker action
type ExecuteResult struct {
	Success bool
	Message string
	Command string
}

// Execute executes a Docker action for the given project
func Execute(actionID, projectPath, projectName string, info *DockerInfo) ExecuteResult {
	switch actionID {
	case "docker-build":
		return dockerBuild(projectPath, projectName, info)
	case "docker-run":
		return dockerRun(projectPath, projectName, false)
	case "docker-run-detached":
		return dockerRun(projectPath, projectName, true)
	case "docker-ps":
		return dockerPS(projectPath)
	case "docker-logs":
		return dockerLogs(projectPath, projectName)
	case "compose-up":
		return composeUp(projectPath, info, false)
	case "compose-up-detached":
		return composeUp(projectPath, info, true)
	case "compose-down":
		return composeDown(projectPath, info)
	case "compose-build":
		return composeBuild(projectPath, info)
	case "compose-logs":
		return composeLogs(projectPath, info)
	case "compose-ps":
		return composePS(projectPath, info)
	default:
		return ExecuteResult{
			Success: false,
			Message: "Unknown Docker action: " + actionID,
		}
	}
}

// dockerBuild builds a Docker image
func dockerBuild(projectPath, projectName string, info *DockerInfo) ExecuteResult {
	dockerfile := GetPrimaryDockerfile(info.Dockerfiles)
	imageName := sanitizeImageName(projectName)

	args := []string{"build", "-t", imageName}
	if dockerfile != "Dockerfile" {
		args = append(args, "-f", dockerfile)
	}
	args = append(args, ".")

	cmd := exec.Command("docker", args...)
	cmd.Dir = projectPath

	output, err := runCommand(cmd)
	cmdStr := fmt.Sprintf("docker %s", strings.Join(args, " "))

	if err != nil {
		return ExecuteResult{
			Success: false,
			Message: fmt.Sprintf("Build failed:\n%s", output),
			Command: cmdStr,
		}
	}

	return ExecuteResult{
		Success: true,
		Message: fmt.Sprintf("Successfully built image: %s\n\n%s", imageName, output),
		Command: cmdStr,
	}
}

// dockerRun runs a Docker container
func dockerRun(projectPath, projectName string, detached bool) ExecuteResult {
	imageName := sanitizeImageName(projectName)

	args := []string{"run"}
	if detached {
		args = append(args, "-d")
	} else {
		args = append(args, "-it", "--rm")
	}
	args = append(args, "--name", imageName+"-container", imageName)

	cmd := exec.Command("docker", args...)
	cmd.Dir = projectPath

	cmdStr := fmt.Sprintf("docker %s", strings.Join(args, " "))

	if detached {
		output, err := runCommand(cmd)
		if err != nil {
			return ExecuteResult{
				Success: false,
				Message: fmt.Sprintf("Failed to run container:\n%s", output),
				Command: cmdStr,
			}
		}
		return ExecuteResult{
			Success: true,
			Message: fmt.Sprintf("Container started: %s", strings.TrimSpace(output)),
			Command: cmdStr,
		}
	}

	// For interactive mode, we need to return the command to be executed by the shell
	return ExecuteResult{
		Success: true,
		Message: fmt.Sprintf("Run this command in your terminal:\ncd %s && %s", projectPath, cmdStr),
		Command: cmdStr,
	}
}

// dockerPS lists running containers
func dockerPS(projectPath string) ExecuteResult {
	cmd := exec.Command("docker", "ps", "--format", "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}")
	cmd.Dir = projectPath

	output, err := runCommand(cmd)
	cmdStr := "docker ps"

	if err != nil {
		return ExecuteResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list containers:\n%s", output),
			Command: cmdStr,
		}
	}

	if strings.TrimSpace(output) == "" || strings.Count(output, "\n") <= 1 {
		return ExecuteResult{
			Success: true,
			Message: "No running containers",
			Command: cmdStr,
		}
	}

	return ExecuteResult{
		Success: true,
		Message: output,
		Command: cmdStr,
	}
}

// dockerLogs shows container logs
func dockerLogs(projectPath, projectName string) ExecuteResult {
	containerName := sanitizeImageName(projectName) + "-container"

	cmd := exec.Command("docker", "logs", "--tail", "50", containerName)
	cmd.Dir = projectPath

	output, err := runCommand(cmd)
	cmdStr := fmt.Sprintf("docker logs %s", containerName)

	if err != nil {
		return ExecuteResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get logs:\n%s", output),
			Command: cmdStr,
		}
	}

	return ExecuteResult{
		Success: true,
		Message: output,
		Command: cmdStr,
	}
}

// composeUp starts services with docker compose
func composeUp(projectPath string, info *DockerInfo, detached bool) ExecuteResult {
	composeFile := GetPrimaryComposeFile(info.ComposeFiles)

	args := []string{"compose", "-f", composeFile, "up"}
	if detached {
		args = append(args, "-d")
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = projectPath

	output, err := runCommand(cmd)
	cmdStr := fmt.Sprintf("docker %s", strings.Join(args, " "))

	if err != nil {
		return ExecuteResult{
			Success: false,
			Message: fmt.Sprintf("Compose up failed:\n%s", output),
			Command: cmdStr,
		}
	}

	return ExecuteResult{
		Success: true,
		Message: fmt.Sprintf("Services started:\n%s", output),
		Command: cmdStr,
	}
}

// composeDown stops and removes services
func composeDown(projectPath string, info *DockerInfo) ExecuteResult {
	composeFile := GetPrimaryComposeFile(info.ComposeFiles)

	args := []string{"compose", "-f", composeFile, "down"}

	cmd := exec.Command("docker", args...)
	cmd.Dir = projectPath

	output, err := runCommand(cmd)
	cmdStr := fmt.Sprintf("docker %s", strings.Join(args, " "))

	if err != nil {
		return ExecuteResult{
			Success: false,
			Message: fmt.Sprintf("Compose down failed:\n%s", output),
			Command: cmdStr,
		}
	}

	return ExecuteResult{
		Success: true,
		Message: fmt.Sprintf("Services stopped:\n%s", output),
		Command: cmdStr,
	}
}

// composeBuild builds all images
func composeBuild(projectPath string, info *DockerInfo) ExecuteResult {
	composeFile := GetPrimaryComposeFile(info.ComposeFiles)

	args := []string{"compose", "-f", composeFile, "build"}

	cmd := exec.Command("docker", args...)
	cmd.Dir = projectPath

	output, err := runCommand(cmd)
	cmdStr := fmt.Sprintf("docker %s", strings.Join(args, " "))

	if err != nil {
		return ExecuteResult{
			Success: false,
			Message: fmt.Sprintf("Compose build failed:\n%s", output),
			Command: cmdStr,
		}
	}

	return ExecuteResult{
		Success: true,
		Message: fmt.Sprintf("Images built:\n%s", output),
		Command: cmdStr,
	}
}

// composeLogs shows logs for all services
func composeLogs(projectPath string, info *DockerInfo) ExecuteResult {
	composeFile := GetPrimaryComposeFile(info.ComposeFiles)

	args := []string{"compose", "-f", composeFile, "logs", "--tail", "50"}

	cmd := exec.Command("docker", args...)
	cmd.Dir = projectPath

	output, err := runCommand(cmd)
	cmdStr := fmt.Sprintf("docker %s", strings.Join(args, " "))

	if err != nil {
		return ExecuteResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get logs:\n%s", output),
			Command: cmdStr,
		}
	}

	return ExecuteResult{
		Success: true,
		Message: output,
		Command: cmdStr,
	}
}

// composePS lists services
func composePS(projectPath string, info *DockerInfo) ExecuteResult {
	composeFile := GetPrimaryComposeFile(info.ComposeFiles)

	args := []string{"compose", "-f", composeFile, "ps"}

	cmd := exec.Command("docker", args...)
	cmd.Dir = projectPath

	output, err := runCommand(cmd)
	cmdStr := fmt.Sprintf("docker %s", strings.Join(args, " "))

	if err != nil {
		return ExecuteResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list services:\n%s", output),
			Command: cmdStr,
		}
	}

	return ExecuteResult{
		Success: true,
		Message: output,
		Command: cmdStr,
	}
}

// runCommand executes a command and returns combined stdout/stderr
func runCommand(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
}

// sanitizeImageName converts a project name to a valid Docker image name
func sanitizeImageName(name string) string {
	// Docker image names must be lowercase and can contain:
	// - lowercase letters, digits, dots, underscores, hyphens
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	// Remove any other invalid characters
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			result += string(r)
		}
	}
	return result
}

// GetActionsForProject returns the available Docker actions for a project
func GetActionsForProject(info *DockerInfo) []Action {
	available := []Action{}

	for _, action := range AvailableActions {
		// Check if action requirements are met
		if action.RequireFile && !info.HasDockerfile {
			continue
		}
		if action.RequireComp && !info.HasCompose {
			continue
		}
		available = append(available, action)
	}

	return available
}
