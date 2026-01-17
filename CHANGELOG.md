# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Docker Integration** - Comprehensive Docker and Docker Compose support
  - Automatic detection of Dockerfile (including variants like Dockerfile.dev, Dockerfile.prod)
  - Automatic detection of Docker Compose files (docker-compose.yml, compose.yml, etc.)
  - Visual indicators in project list (🐳 for Dockerfile, 🐙 for Compose)
  - Docker actions in action menu:
    - Build Docker images
    - Run containers (interactive and detached)
    - View container logs
    - List running containers
    - Docker Compose up/down operations
    - Compose build, logs, and service listing
  - Smart primary file selection for projects with multiple Dockerfiles
  - Automatic image name sanitization
  - Dev container detection (.devcontainer/devcontainer.json)
  - See [docs/DOCKER.md](docs/DOCKER.md) for full documentation

### Changed
- Project struct now includes `HasDockerfile` and `HasCompose` boolean fields
- Action menu dynamically shows Docker actions based on available files
- Project list view displays Docker indicators alongside git and language info

### Technical
- Added `internal/docker` package with detection and action modules
- Comprehensive test suite for Docker functionality (100% passing)
- Integration with existing action executor system
- No breaking changes to existing functionality

## Future Releases

### Planned Features
- Real-time container status indicators
- Container health monitoring
- Docker registry integration
- Kubernetes awareness
- Docker-specific configuration options

---

## How to Use This Changelog

- **[Unreleased]** - Features merged but not yet released
- **Version numbers** - Released versions with dates
- **Added** - New features
- **Changed** - Changes to existing functionality
- **Deprecated** - Features that will be removed
- **Removed** - Features that were removed
- **Fixed** - Bug fixes
- **Security** - Security-related changes
