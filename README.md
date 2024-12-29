## Ollama Notes
1. Ollama calls won't work without the port forward command:
`kubectl -n ollama port-forward service/ollama 11434:80`
2. Models aren't automatically loaded into the image. If the image restarts we need to reload the llama3.2:3b image. Exec into the ollama pod and run:
`ollama run llama3.2:3b`

## TODO
1. Create a new image of the running ollama container.
2. Create a web interface to interact and receive output.
3. Create new models.

