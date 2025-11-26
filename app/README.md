# Demo App

A simple Go web application built with Gorilla Mux for demonstrating Kubernetes deployments with Flagger canary releases.

## Features

- RESTful API endpoints
- Health check endpoints
- Version information
- Runtime and environment details
- JSON responses

## API Endpoints

- `GET /` - Home endpoint, returns application information
- `GET /healthz` - Health check endpoint
- `GET /readyz` - Readiness check endpoint
- `GET /version` - Version information

## Building

```bash
docker build -t demo-app:latest .
```

## Running Locally

```bash
# Using Go
go run main.go

# Using Docker
docker run -p 9898:9898 demo-app:latest
```

The application will be available at `http://localhost:9898`

## Environment Variables

- `PORT` - Server port (default: 9898)
- `REVISION` - Git revision or version identifier
- `COLOR` - UI color theme (default: #34577c)
- `ENVIRONMENT` - Environment name (default: unknown)
- `LOG_LEVEL` - Logging level (default: info)

## Example Response

```json
{
  "hostname": "demo-app-746d5c88bd-8c4cl",
  "version": "1.0.0",
  "revision": "abc123",
  "color": "#34577c",
  "message": "Hello from Flagger Demo!",
  "runtime": "go1.21.13 linux/arm64",
  "uptime": "5m20.103273119s",
  "env": {
    "ENVIRONMENT": "production",
    "LOG_LEVEL": "info"
  }
}
```
