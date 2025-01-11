## Ollama Notes
# Local
1. Ollama calls won't work without the port forward command:
`kubectl -n ollama port-forward service/ollama 11434:80`
2. Rename app.py.local and app.py so that app.py local is able to run with command
`python app.py`
# K8s
1. You can use my deployment with my Dockerhub registry but it may be better to create and upload the Docker image to your own repo.

## TODO
1. Create a web interface to interact and receive output.
2. Create new models.

