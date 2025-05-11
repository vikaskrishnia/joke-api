# Joke API

A simple RESTful API that serves jokes in multiple languages (English, Spanish, French, German, and Hindi).

## Features

- Get random jokes in different languages
- Prometheus metrics for monitoring
- Docker/Kubernetes ready
- Built with Go

## API Endpoints

- `/joke` - Get a random joke (default: English)
  - Query parameters:
    - `lang` - Language code (en, es, fr, de, hi)
- `/metrics` - Prometheus metrics endpoint

## Running with Docker

```bash
docker build -t joke-api .
docker run -p 8080:8080 joke-api
```

## Running with Podman

```bash
podman build -t joke-api .
podman run -p 8080:8080 joke-api
```

## License

MIT
