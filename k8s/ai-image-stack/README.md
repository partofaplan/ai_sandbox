# AI Image Stack (Generator + Gallery)

This folder contains two small apps and Kubernetes manifests that share a PVC:
- **Image Generator UI**: Gradio + Diffusers text-to-image app.
- **Image Gallery UI**: FastAPI app that browses the same output directory.

## Build images

```bash
docker build -t local/ai-image-gen-ui:latest k8s/ai-image-stack/apps/image-gen-ui
docker build -t local/ai-image-gallery:latest k8s/ai-image-stack/apps/image-gallery
```

If you're using a local cluster (kind/k3d), load images into the cluster:

```bash
# kind
kind load docker-image local/ai-image-gen-ui:latest
kind load docker-image local/ai-image-gallery:latest

# k3d
k3d image import local/ai-image-gen-ui:latest
k3d image import local/ai-image-gallery:latest
```

## Deploy with Helm

```bash
helm upgrade --install ai-image-stack k8s/ai-image-stack/chart/ai-image-stack \
  --namespace ai-image-stack --create-namespace
```

The Helm chart defaults to a Traefik ingress class and example hosts:
- `image-gen.local`
- `image-gallery.local`

If you need to override images:

```bash
helm upgrade --install ai-image-stack k8s/ai-image-stack/chart/ai-image-stack \
  --namespace ai-image-stack --create-namespace \
  --set imageGen.image=partofaplan/ai-image-gen-ui \
  --set imageGen.tag=latest \
  --set gallery.image=partofaplan/ai-image-gallery \
  --set gallery.tag=latest
```

## Deploy with Kustomize

```bash
kubectl apply -k k8s/ai-image-stack
```

## Access the UIs

### Option A: Port-forward

If you used Helm with the release name `ai-image-stack`:

```bash
kubectl -n ai-image-stack port-forward svc/ai-image-stack-gen-ui 7860:80
kubectl -n ai-image-stack port-forward svc/ai-image-stack-gallery 8081:80
```

If you used Kustomize:

```bash
kubectl -n ai-image-stack port-forward svc/ai-image-gen-ui 7860:80
kubectl -n ai-image-stack port-forward svc/ai-image-gallery 8081:80
```

- Generator UI: http://localhost:7860
- Gallery UI: http://localhost:8081

### Option B: Ingress

Edit `k8s/ai-image-stack/chart/ai-image-stack/values.yaml` (Helm) or
`k8s/ai-image-stack/manifests/ingress.yaml` (Kustomize) to match your ingress class and hostnames.
Example hosts are:
- `image-gen.local`
- `image-gallery.local`

Add them to your `/etc/hosts` and visit in a browser.

## Notes

- Images are stored in `/data/outputs` on the shared PVC (`ai-image-stack-data`).
- The PVC uses `ReadWriteOnce`. If your cluster schedules pods to different nodes, either use an RWX-capable storage class or add node affinity to keep both pods on the same node.
- Model weights are cached under `/data/hf` so they persist across restarts.
- To use a GPU base image, swap the generator Dockerfile base image (e.g. `pytorch/pytorch:2.4.0-cuda11.8-cudnn8-runtime`) and add GPU resources/limits in the deployment.
