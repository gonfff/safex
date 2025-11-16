# Docker Deployment

This guide covers deploying Safex using Docker and Docker Compose.

## Quick Start with Docker

### Using Pre-built Image

Pull and run the latest image from GitHub Container Registry:

```bash
# Pull the image
docker pull ghcr.io/gonfff/safex:latest

# Run with local storage
docker run -p 8000:8000 ghcr.io/gonfff/safex:latest
```

### Building from Source

Build the Docker image locally:

```bash
# Clone the repository
git clone https://github.com/gonfff/safex.git
cd safex

# Build the image
docker build -t safex:local .

# Run locally built image
docker run -p 8000:8000 safex:local
```
