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
   kubectl port-forward service/zachbot-zachbot-ollama 11434:80
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
| `OLLAMA_RELOAD_NAME` | `prospector` | Model name to rebuild when reloading     |
| `OLLAMA_RELOAD_PATH` | `/root/models/Modelfile` | Path to the Modelfile in the Ollama container |
| `CORS_ALLOW_ORIGINS` | `*;http://localhost:6600` | Semi-colon or comma separated allow list for CORS |
| `STATIC_DIR` | `static` | Directory from which to serve static assets |
| `TEMPLATE_DIR` | `templates` | Directory from which to load HTML templates |

### Penpal (email bot)

Environment variables for the penpal service (defaults are set in the Helm chart):

| Variable | Default | Description |
|----------|---------|-------------|
| `PENPAL_MAILHOG_API` | `http://mailhog:8025` | Mailhog API endpoint for fetching inbound mail |
| `PENPAL_SMTP_HOST` | `mailhog` | SMTP host used to send replies |
| `PENPAL_SMTP_PORT` | `1025` | SMTP port |
| `PENPAL_FROM` | `penpal@zachbot.local` | From address used for replies |
| `PENPAL_POLL_INTERVAL` | `15s` | How often to poll for new messages |
| `PENPAL_MODEL` | `prospector:latest` | Default model used for replies |
| `PENPAL_PERSONA` | `Write a concise, friendly email reply.` | Default persona injected into replies |

## Helm deployment

A Helm chart lives under `k8s/helm/zachbot` and deploys Zachbot alongside an Ollama instance.

```sh
helm lint k8s/helm/zachbot
helm template test-release k8s/helm/zachbot
# helm install zachbot k8s/helm/zachbot
```

Override values as needed, for example to use custom images:

```sh
helm install zachbot k8s/helm/zachbot \
  --set zachbot.image.repository=myrepo/zachbot \
  --set ollama.image.repository=myrepo/ollama
```

The chart also provisions Traefik ingresses for both services. By default,
Zachbot is available at `http://zachbot.localhost` and the Ollama API at
`http://ollama.localhost`. These hosts correspond to the
`zachbot.ingress.host` and `ollama.ingress.host` entries in `values.yaml` and
may be overridden as needed.

## Project structure

```
go_webapp/                 # Go source code for Zachbot
k8s/
  helm/
    zachbot/               # Helm chart deploying Zachbot and Ollama
  legacy/                  # Raw Kubernetes manifests
legacy/                    # Older experiments
```

## Development

Run tests before submitting changes:

```sh
go test ./...
```

## TODO

- Create a web interface to interact and receive output
- Create new models
