# Dispatcherd

Dispatcherd is a message dispatching service that routes messages based on configurable rules to various dispatchers such as log files, counters, or email.

## Features

- Rule-based message routing
- Multiple dispatcher types (log, counter, email)
- REST API for message submission
- Configurable through file-based configurations
- Docker support for easy deployment

## Prerequisites

- Go 1.25 or higher
- Docker (for containerized deployment)
- Task (https://taskfile.dev/) - optional but recommended

## Building the Application

### Using Task (Recommended)

```bash
# Build the application
task build

# Run the application
task run
```

### Using Go Directly

```bash
# Build the application
go build -o build/dispatcherd ./cmd/

# Run the application
./build/dispatcherd
```

## Running with Docker

### Building the Docker Image

```bash
docker build -t dispatcherd .
```

### Running the Docker Container

```bash
docker run -p 3001:3001 \
  -v /path/to/rules:/data/rules \
  -v /path/to/dispatchers:/data/dispatchers \
  -e DISPATCHERD_LISTEN_ADDRESS=:3001 \
  -e DISPATCHERD_RULE_DIRECTORY=/data/rules \
  -e DISPATCHERD_DISPATCHER_CONFIG_DIRECTORY=/data/dispatchers \
  dispatcherd
```

### Docker Compose Example

Create a `docker-compose.yml` file:

```yaml
version: '3.8'
services:
  dispatcherd:
    build: .
    ports:
      - "3001:3001"
    volumes:
      - ./rules:/data/rules
      - ./dispatchers:/data/dispatchers
    environment:
      - DISPATCHERD_LISTEN_ADDRESS=:3001
      - DISPATCHERD_RULE_DIRECTORY=/data/rules
      - DISPATCHERD_DISPATCHER_CONFIG_DIRECTORY=/data/dispatchers
```

Then run:

```bash
docker-compose up
```

## Configuration

Dispatcherd is configured through environment variables and file-based configurations.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| DISPATCHERD_LISTEN_ADDRESS | Address to listen on | :3001 |
| DISPATCHERD_LOG_LEVEL | Log level (DEBUG, INFO, WARN, ERROR) | DEBUG |
| DISPATCHERD_ENVIRONMENT | Environment (dev, prod) | prod |
| DISPATCHERD_CORS_ALLOWED_ORIGIN | Allowed CORS origin | * |
| DISPATCHERD_RULE_DIRECTORY | Directory containing rule files | /data/rules |
| DISPATCHERD_DISPATCHER_CONFIG_DIRECTORY | Directory containing dispatcher config files | /data/dispatchers |

### Rule Configuration

Rules are defined in JSON files in the rules directory. Each file should contain a single rule object:

```json
{
  "id": "log when type tag is log",
  "dispatcherName": "log-error",
  "match": [
    {
      "tagName": "type",
      "operator": "eq",
      "value": "log"
    }
  ]
}
```

### Dispatcher Configuration

Dispatcher configurations are defined in JSON files in the dispatchers directory:

```json
{
  "name": "log-error",
  "type": "log",
  "config": {
    "level": 8
  }
}
```

## API Endpoints

- `POST /message` - Submit a message for dispatching
- `GET /health` - Health check endpoint

## Development

### Testing

```bash
# Run tests
task test
```

### Linting

```bash
# Run linter
task lint

# Fix linting issues
task lint:fix
```
