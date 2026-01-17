package language

import (
	"os"
	"path/filepath"
	"strings"
)

// Detector defines a language detector
type Detector struct {
	Language string
	Priority int
	Check    func(files []string) bool
}

// detectors holds all language detectors in priority order
var detectors = []Detector{
	{
		Language: "Go",
		Priority: 1,
		Check: func(files []string) bool {
			return contains(files, "go.mod") || contains(files, "go.sum")
		},
	},
	{
		Language: "Rust",
		Priority: 2,
		Check: func(files []string) bool {
			return contains(files, "Cargo.toml")
		},
	},
	{
		Language: "TypeScript",
		Priority: 3,
		Check: func(files []string) bool {
			return contains(files, "tsconfig.json") || hasExtension(files, ".ts", ".tsx")
		},
	},
	{
		Language: "JavaScript",
		Priority: 4,
		Check: func(files []string) bool {
			return contains(files, "package.json")
		},
	},
	{
		Language: "Python",
		Priority: 5,
		Check: func(files []string) bool {
			return contains(files, "pyproject.toml") || contains(files, "requirements.txt") ||
				contains(files, "setup.py") || contains(files, "Pipfile")
		},
	},
	{
		Language: "Java",
		Priority: 6,
		Check: func(files []string) bool {
			return contains(files, "pom.xml") || contains(files, "build.gradle") ||
				contains(files, "build.gradle.kts")
		},
	},
	{
		Language: "C#",
		Priority: 7,
		Check: func(files []string) bool {
			return hasExtension(files, ".csproj", ".sln")
		},
	},
	{
		Language: "Ruby",
		Priority: 8,
		Check: func(files []string) bool {
			return contains(files, "Gemfile") || hasExtension(files, ".gemspec")
		},
	},
	{
		Language: "PHP",
		Priority: 9,
		Check: func(files []string) bool {
			return contains(files, "composer.json")
		},
	},
	{
		Language: "Swift",
		Priority: 10,
		Check: func(files []string) bool {
			return contains(files, "Package.swift") || hasExtension(files, ".swift")
		},
	},
	{
		Language: "Kotlin",
		Priority: 11,
		Check: func(files []string) bool {
			return hasExtension(files, ".kt", ".kts")
		},
	},
	{
		Language: "C/C++",
		Priority: 12,
		Check: func(files []string) bool {
			hasSourceFiles := hasExtension(files, ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp")
			hasBuildFiles := contains(files, "Makefile") || contains(files, "CMakeLists.txt")
			return hasSourceFiles && hasBuildFiles
		},
	},
	{
		Language: "Elixir",
		Priority: 13,
		Check: func(files []string) bool {
			return contains(files, "mix.exs")
		},
	},
	{
		Language: "Zig",
		Priority: 14,
		Check: func(files []string) bool {
			return contains(files, "build.zig")
		},
	},
	{
		Language: "Haskell",
		Priority: 15,
		Check: func(files []string) bool {
			return contains(files, "stack.yaml") || hasExtension(files, ".cabal")
		},
	},
	{
		Language: "Scala",
		Priority: 16,
		Check: func(files []string) bool {
			return contains(files, "build.sbt")
		},
	},
	{
		Language: "Clojure",
		Priority: 17,
		Check: func(files []string) bool {
			return contains(files, "project.clj") || contains(files, "deps.edn")
		},
	},
}

// Detect detects the primary language of a project
func Detect(projectPath string) (string, error) {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return "Unknown", err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	// Try each detector in priority order
	for _, detector := range detectors {
		if detector.Check(files) {
			return detector.Language, nil
		}
	}

	// Check if it's at least a git repo
	if contains(files, ".git") {
		return "Git Repo", nil
	}

	return "Unknown", nil
}

// GetIcon returns an icon for the language
func GetIcon(language string) string {
	icons := map[string]string{
		"Go":         "🐹",
		"Rust":       "🦀",
		"TypeScript": "⚡",
		"JavaScript": "📜",
		"Python":     "🐍",
		"Java":       "☕",
		"C#":         "🔷",
		"Ruby":       "💎",
		"PHP":        "🐘",
		"Swift":      "🦅",
		"Kotlin":     "🎯",
		"C/C++":      "⚙️",
		"Elixir":     "💧",
		"Zig":        "⚡",
		"Haskell":    "λ",
		"Scala":      "🔺",
		"Clojure":    "🔄",
		"Git Repo":   "📁",
		"Unknown":    "❓",
	}

	if icon, ok := icons[language]; ok {
		return icon
	}
	return "📄"
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// hasExtension checks if any file has one of the given extensions
func hasExtension(files []string, extensions ...string) bool {
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		for _, targetExt := range extensions {
			if ext == strings.ToLower(targetExt) {
				return true
			}
		}
	}
	return false
}

// AddDetector adds a custom language detector
func AddDetector(detector Detector) {
	detectors = append(detectors, detector)
	// Could sort by priority here if needed
}

// GetDetectors returns all registered detectors
func GetDetectors() []Detector {
	return detectors
}
