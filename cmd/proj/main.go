package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/s33g/proj/internal/app"
	"github.com/s33g/proj/internal/config"
	"github.com/s33g/proj/internal/project"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("proj version %s\n", version)
			return

		case "--help", "-h":
			printHelp()
			return

		case "--init":
			if err := initConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		case "--config":
			if err := openConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		case "--set-path":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Error: --set-path requires a path argument")
				os.Exit(1)
			}
			if err := setPath(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		case "--list", "-l":
			if err := listProjects(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		default:
			// Check if it's a project name
			if !strings.HasPrefix(os.Args[1], "-") {
				if err := jumpToProject(os.Args[1]); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			}
			fmt.Fprintf(os.Stderr, "Unknown option: %s\n", os.Args[1])
			printHelp()
			os.Exit(1)
		}
	}

	// Run the TUI
	if err := runTUI(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`proj - TUI Project Navigator

Usage:
  proj                    Launch TUI
  proj <project-name>     Jump directly to project
  proj --list             List all projects (non-interactive)
  proj --init             Initialize/reset configuration
  proj --config           Open config in $EDITOR
  proj --set-path <path>  Set projects directory
  proj --version          Show version
  proj --help             Show help

Keyboard shortcuts (in TUI):
  Enter   Select item
  /       Search/filter
  q       Quit
  Esc     Back/Cancel

Documentation: https://github.com/s33g/proj`)
}

func initConfig() error {
	cfg := config.DefaultConfig()

	configDir, err := config.ConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Configuration initialized at %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Set your projects directory:")
	fmt.Println("     proj --set-path ~/code")
	fmt.Println("  2. Launch the TUI:")
	fmt.Println("     proj")
	return nil
}

func openConfig() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	configDir, err := config.ConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found, run 'proj --init' first")
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func setPath(path string) error {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	cfg.ReposPath = absPath

	// Save config
	configDir, err := config.ConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Projects path set to: %s\n", absPath)
	return nil
}

func listProjects() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w (run 'proj --init' first)", err)
	}

	scanner := project.NewScanner(cfg)
	projects, err := scanner.Scan(cfg.ReposPath)
	if err != nil {
		return fmt.Errorf("failed to scan projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	for _, p := range projects {
		fmt.Println(p.Name)
	}
	return nil
}

func jumpToProject(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	scanner := project.NewScanner(cfg)
	projects, err := scanner.Scan(cfg.ReposPath)
	if err != nil {
		return fmt.Errorf("failed to scan projects: %w", err)
	}

	// Find matching project (case-insensitive partial match)
	var match *project.Project
	nameLower := strings.ToLower(name)
	for _, p := range projects {
		if strings.ToLower(p.Name) == nameLower {
			match = p
			break
		}
		if strings.Contains(strings.ToLower(p.Name), nameLower) && match == nil {
			match = p
		}
	}

	if match == nil {
		return fmt.Errorf("project not found: %s", name)
	}

	// Write path to cd file if set
	cdFile := os.Getenv("PROJ_CD_FILE")
	if cdFile != "" {
		if err := os.WriteFile(cdFile, []byte(match.Path), 0644); err != nil {
			return fmt.Errorf("failed to write cd file: %w", err)
		}
	} else {
		// Just print the path
		fmt.Println(match.Path)
	}

	return nil
}

func runTUI() error {
	cfg, err := config.Load()
	if err != nil {
		// If no config exists, run init flow
		fmt.Println("Welcome to proj! Let's set up your configuration.")
		fmt.Println()
		if err := initConfig(); err != nil {
			return err
		}
		return nil
	}

	model := app.New(cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	// Handle post-exit actions
	if m, ok := finalModel.(app.Model); ok {
		// Handle cd path
		cdPath := m.GetCdPath()
		if cdPath != "" {
			cdFile := os.Getenv("PROJ_CD_FILE")
			if cdFile != "" {
				if err := os.WriteFile(cdFile, []byte(cdPath), 0644); err != nil {
					return fmt.Errorf("failed to write cd file: %w", err)
				}
			}
		}

		// Handle exec command
		execCmd := m.GetExecCmd()
		if len(execCmd) > 0 {
			binary, err := exec.LookPath(execCmd[0])
			if err != nil {
				return fmt.Errorf("command not found: %s", execCmd[0])
			}
			// Replace current process with the command
			return syscall.Exec(binary, execCmd, os.Environ())
		}
	}

	return nil
}
