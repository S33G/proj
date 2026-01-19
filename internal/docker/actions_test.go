package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeImageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase preserved", "simple-name", "simple-name"},
		{"spaces to hyphens", "My App", "my-app"},
		{"invalid chars removed", "App@#Name!", "appname"},
		{"dots underscores kept", "app.Name_v1", "app.name_v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeImageName(tt.input); got != tt.expected {
				t.Fatalf("sanitizeImageName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetActionsForProjectFiltersByCapabilities(t *testing.T) {
	info := &DockerInfo{HasDockerfile: true, HasCompose: true}
	actions := GetActionsForProject(info)

	if len(actions) != len(AvailableActions) {
		t.Fatalf("expected all actions when both capabilities present, got %d", len(actions))
	}

	info = &DockerInfo{HasDockerfile: false, HasCompose: false}
	actions = GetActionsForProject(info)
	for _, action := range actions {
		if action.RequireFile || action.RequireComp {
			t.Fatalf("unexpected action requiring missing capability: %+v", action)
		}
	}
}

func TestDockerPSOutputHandling(t *testing.T) {
	t.Run("reports running containers", func(t *testing.T) {
		withFakeDocker(t, `#!/bin/sh
if [ "$1" = "ps" ]; then
  printf "CONTAINER ID\tIMAGE\tSTATUS\tNAMES\n123\timg\tUp\tname\n"
  exit 0
fi
echo "unexpected" >&2
exit 1
`, func() {
			result := dockerPS(t.TempDir())
			if !result.Success {
				t.Fatalf("expected success, got failure: %s", result.Message)
			}
			if !strings.Contains(result.Message, "123") {
				t.Fatalf("expected container output, got %q", result.Message)
			}
		})
	})

	t.Run("handles empty list", func(t *testing.T) {
		withFakeDocker(t, `#!/bin/sh
if [ "$1" = "ps" ]; then
  printf "CONTAINER ID\tIMAGE\tSTATUS\tNAMES\n"
  exit 0
fi
exit 1
`, func() {
			result := dockerPS(t.TempDir())
			if !result.Success {
				t.Fatalf("expected success for empty ps output, got failure: %s", result.Message)
			}
			if result.Message != "No running containers" {
				t.Fatalf("expected empty ps message, got %q", result.Message)
			}
		})
	})

	t.Run("propagates failures", func(t *testing.T) {
		withFakeDocker(t, `#!/bin/sh
echo "boom" >&2
exit 1
`, func() {
			result := dockerPS(t.TempDir())
			if result.Success {
				t.Fatalf("expected failure, got success: %s", result.Message)
			}
			if !strings.Contains(result.Message, "boom") {
				t.Fatalf("expected error output in message, got %q", result.Message)
			}
		})
	})
}

func withFakeDocker(t *testing.T, script string, fn func()) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "docker")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake docker: %v", err)
	}

	origPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath); err != nil {
		t.Fatalf("failed to set PATH: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("PATH", origPath)
	})

	fn()
}
