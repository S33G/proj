# Docker Integration

The `proj` TUI now includes comprehensive Docker and Docker Compose support, enabling you to manage containerized projects directly from the interface.

## Features

### 🐳 Automatic Detection

`proj` automatically detects Docker-related files in your projects:

- **Dockerfile** - Including variants like `Dockerfile.dev`, `Dockerfile.prod`
- **Docker Compose** - Supports `docker-compose.yml`, `docker-compose.yaml`, `compose.yml`, `compose.yaml`
- **Dev Containers** - Detects `.devcontainer/devcontainer.json`

### 📊 Visual Indicators

Projects with Docker support display visual indicators in the project list:

```
🐹 my-go-api          main   🐳    (has Dockerfile)
🐍 ml-pipeline        dev    🐙    (has Docker Compose)
⚡ web-frontend       feat   
```

- 🐳 = Has Dockerfile
- 🐙 = Has Docker Compose (takes precedence over Dockerfile icon)

### 🎯 Docker Actions

When you select a project with Docker support, the following actions become available:

#### Single Container Actions (Dockerfile)

| Action | Description | Command |
|--------|-------------|---------|
| 🏗️ Build Image | Build Docker image from Dockerfile | `docker build -t <project-name> .` |
| ▶️ Run Container | Run container interactively | `docker run -it --rm <project-name>` |
| 🔄 Run Detached | Run container in background | `docker run -d <project-name>` |
| 📋 List Containers | Show running containers | `docker ps` |
| 📜 View Logs | Stream container logs | `docker logs <container-name>` |

#### Docker Compose Actions

| Action | Description | Command |
|--------|-------------|---------|
| 🚀 Compose Up | Start all services | `docker compose up` |
| 🔄 Compose Up (Detached) | Start services in background | `docker compose up -d` |
| 🛑 Compose Down | Stop and remove services | `docker compose down` |
| 🏗️ Compose Build | Build all images | `docker compose build` |
| 📜 Compose Logs | Stream all service logs | `docker compose logs` |
| 📋 Compose PS | List services | `docker compose ps` |

## Usage

### Basic Workflow

1. **Navigate to a project** with Docker support
2. **Press Enter** to view available actions
3. **Select a Docker action** from the menu
4. The action will execute and show results

### Example: Building and Running a Container

```
1. Select project with Dockerfile
2. Press Enter → Select "🏗️ Build Image"
3. Wait for build to complete
4. Press Enter again → Select "▶️ Run Container"
5. Container starts interactively
```

### Example: Managing Compose Services

```
1. Select project with docker-compose.yml
2. Press Enter → Select "🚀 Compose Up (Detached)"
3. Services start in background
4. Select "📋 Compose PS" to view status
5. Select "🛑 Compose Down" to stop when done
```

## Multiple Dockerfile Support

If your project has multiple Dockerfiles (e.g., `Dockerfile`, `Dockerfile.dev`, `Dockerfile.prod`), `proj` will:

1. **Prefer** the plain `Dockerfile` if it exists
2. **Fallback** to the first variant found
3. **Use** the appropriate `-f` flag when building

You can see which Dockerfile will be used in the Docker detection info.

## Multiple Compose File Support

For projects with multiple compose files, the preference order is:

1. `compose.yml` (modern, shorter name)
2. `docker-compose.yml` (traditional)
3. `compose.yaml` (YAML variant)
4. `docker-compose.yaml` (traditional YAML)

## Image Naming

Docker images are automatically named based on your project directory name:

- Project: `my-awesome-app` → Image: `my-awesome-app`
- Project: `My Project` → Image: `my-project` (sanitized)
- Invalid characters are removed or replaced with hyphens

## Requirements

### Docker Installation

You must have Docker installed and accessible in your PATH:

```bash
# Check if Docker is available
docker --version

# Check if Docker Compose v2 plugin is available
docker compose version
```

### Permissions

Ensure your user has permission to run Docker commands:

```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER

# Log out and back in for group changes to take effect
```

## Container Status (Coming Soon)

Future versions will include real-time container status indicators:

```
🐹 my-go-api          main   🐳 ●  (container running)
🐍 ml-pipeline        dev    🐙 ◐  (2/3 services up)
⚡ web-frontend       feat   🐳 ○  (no container)
```

Status symbols:
- `●` Running
- `◐` Partial (some services running)
- `○` Stopped/None
- `!` Error/Unhealthy

## Troubleshooting

### "Docker command not found"

Ensure Docker is installed and in your PATH:

```bash
which docker
# Should output: /usr/bin/docker or similar
```

### "Permission denied" errors

Check Docker permissions:

```bash
# Test Docker access
docker ps

# If permission denied, add user to docker group
sudo usermod -aG docker $USER
```

### Compose v1 vs v2

`proj` uses the modern `docker compose` command (v2). If you have the older `docker-compose` (v1):

**Option 1:** Upgrade to Docker Compose v2
```bash
# Install Docker Compose v2 plugin
# See: https://docs.docker.com/compose/install/
```

**Option 2:** Create an alias
```bash
# Add to ~/.bashrc or ~/.zshrc
alias docker-compose='docker compose'
```

### Actions not showing up

If Docker actions don't appear for your project:

1. Check that Docker files are in the root directory
2. Verify file names match exactly (case-sensitive)
3. Try refreshing the project list

## Configuration (Future)

Future versions will support Docker-specific configuration:

```yaml
# ~/.config/proj/config.yml
docker:
  enabled: true
  composeCommand: "docker compose"  # or "docker-compose" for v1
  defaultBuildArgs: []
  statusRefreshInterval: 30  # seconds
  showContainerStatus: true
```

## Examples

### Simple Node.js App with Dockerfile

```dockerfile
# Dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
```

**Available actions:**
- Build Image
- Run Container
- Run Detached
- List Containers
- View Logs

### Full-Stack App with Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
  backend:
    build: ./backend
    ports:
      - "8080:8080"
  database:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: secret
```

**Available actions:**
- Compose Up
- Compose Up (Detached)
- Compose Down
- Compose Build
- Compose Logs
- Compose PS

## Contributing

Found a bug or have a feature request for Docker integration? Please open an issue on GitHub!

## License

MIT License - See LICENSE file for details
