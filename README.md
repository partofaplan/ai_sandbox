# Zachbot

Zachbot is a Go web service that proxies requests to an [Ollama](https://github.com/ollama/ollama) model. The repository also includes a Helm chart for deploying both Zachbot and a companion Ollama service to Kubernetes.

## Prerequisites

- [Go](https://go.dev/) 1.21+
- [Helm](https://helm.sh/) 3.x
- Access to a Kubernetes cluster for optional Helm deployment
- An Ollama model (defaults to `prospector:latest`)

## Running locally

1. Start an Ollama instance or forward the service if it runs in a cluster:
   ```sh
   kubectl -n ollama port-forward service/ollama 11434:80
   ```
2. Build and run the web application:
   ```sh
   cd go_webapp
   go build ./...
   ./go_webapp
   ```
3. Chat with the service:
   ```sh
   curl -sS -X POST http://localhost:6600/chat/ \
        -H "Content-Type: application/json" \
        -d '{"prompt":"hello"}'
   ```

### Configuration
The application can be configured via environment variables:

| Variable       | Default            | Description                              |
|----------------|--------------------|------------------------------------------|
| `SERVER_PORT`  | `6600`             | Port used by the Zachbot server          |
| `API_TIMEOUT`  | `300s`             | Timeout for requests to Ollama           |
| `OLLAMA_HOST`  | `ollama`           | Hostname of the Ollama service           |
| `OLLAMA_PORT`  | `11434`            | Port of the Ollama service               |
| `OLLAMA_MODEL` | `prospector:latest`| Model name used for chat requests        |

## Helm deployment

A Helm chart lives under `k8s/zachbot-chart` and deploys Zachbot alongside an Ollama instance.

```sh
helm lint k8s/zachbot-chart
helm template test-release k8s/zachbot-chart
# helm install zachbot k8s/zachbot-chart
```

Override values as needed, for example to use custom images:

```sh
helm install zachbot k8s/zachbot-chart \
  --set zachbot.image.repository=myrepo/zachbot \
  --set ollama.image.repository=myrepo/ollama
```

## Project structure

```
go_webapp/   # Go source code for Zachbot
k8s/         # Kubernetes manifests and Helm chart
legacy/      # Older experiments
```

## Development

Run tests before submitting changes:

```sh
go test ./...
```

## TODO

- Create a web interface to interact and receive output
- Create new models

