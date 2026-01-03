import os
import time
from pathlib import Path

from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

OUTPUT_DIR = Path(os.getenv("OUTPUT_DIR", "/data/outputs"))
MAX_IMAGES = int(os.getenv("MAX_IMAGES", "200"))
REFRESH_SECONDS = int(os.getenv("REFRESH_SECONDS", "0"))

OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

app = FastAPI(title="Image Gallery")
app.mount("/images", StaticFiles(directory=str(OUTPUT_DIR)), name="images")

templates = Jinja2Templates(directory="templates")


def _list_images():
    images = []
    for path in OUTPUT_DIR.iterdir():
        if path.is_file() and path.suffix.lower() in {".png", ".jpg", ".jpeg", ".webp"}:
            images.append(path)

    images.sort(key=lambda item: item.stat().st_mtime, reverse=True)

    items = []
    for path in images[:MAX_IMAGES]:
        items.append(
            {
                "name": path.name,
                "url": f"/images/{path.name}",
                "mtime": time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(path.stat().st_mtime)),
            }
        )

    return items


@app.get("/", response_class=HTMLResponse)
def index(request: Request):
    images = _list_images()
    return templates.TemplateResponse(
        "index.html",
        {
            "request": request,
            "images": images,
            "refresh_seconds": REFRESH_SECONDS,
        },
    )


@app.get("/healthz")
def healthz():
    return {"status": "ok"}
