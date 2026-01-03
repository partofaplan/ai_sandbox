import os
import time
import uuid
from pathlib import Path

import gradio as gr
import torch
from diffusers import AutoPipelineForText2Image

MODEL_ID = os.getenv("MODEL_ID", "stabilityai/sd-turbo")
MODEL_VARIANT = os.getenv("MODEL_VARIANT", "").strip()
OUTPUT_DIR = Path(os.getenv("OUTPUT_DIR", "/data/outputs"))
MAX_IMAGE_SIDE = int(os.getenv("MAX_IMAGE_SIDE", "1024"))

OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

_pipeline = None


def _get_device() -> str:
    if torch.cuda.is_available():
        return "cuda"
    if torch.backends.mps.is_available():
        return "mps"
    return "cpu"


def _load_pipeline():
    global _pipeline
    if _pipeline is not None:
        return _pipeline

    device = _get_device()
    dtype = torch.float16 if device == "cuda" else torch.float32

    kwargs = {"torch_dtype": dtype}
    if MODEL_VARIANT:
        kwargs["variant"] = MODEL_VARIANT

    _pipeline = AutoPipelineForText2Image.from_pretrained(MODEL_ID, **kwargs)
    _pipeline = _pipeline.to(device)
    _pipeline.enable_attention_slicing()

    return _pipeline


def _normalize_side(value: int) -> int:
    value = max(64, min(MAX_IMAGE_SIDE, int(value)))
    return value - (value % 8)


def generate_image(prompt, negative_prompt, steps, guidance, width, height, seed):
    if not prompt:
        raise gr.Error("Prompt cannot be empty.")

    pipeline = _load_pipeline()
    width = _normalize_side(width)
    height = _normalize_side(height)

    seed = int(seed)
    if seed < 0:
        seed = int(time.time())

    generator = torch.Generator(device=pipeline.device).manual_seed(seed)

    result = pipeline(
        prompt=prompt,
        negative_prompt=negative_prompt or None,
        num_inference_steps=int(steps),
        guidance_scale=float(guidance),
        width=width,
        height=height,
        generator=generator,
    )

    image = result.images[0]
    filename = f"{time.strftime('%Y%m%d-%H%M%S')}-{uuid.uuid4().hex[:8]}.png"
    output_path = OUTPUT_DIR / filename
    image.save(output_path)

    return image, str(output_path)


with gr.Blocks(title="AI Image Generator") as demo:
    gr.Markdown(
        "# AI Image Generator\n"
        "Generate images from text prompts. Outputs are saved to the shared PVC."
    )

    with gr.Row():
        prompt = gr.Textbox(label="Prompt", placeholder="A neon skyline at dusk, cinematic lighting")
        negative_prompt = gr.Textbox(label="Negative prompt", placeholder="blurry, distorted")

    with gr.Row():
        steps = gr.Slider(1, 50, value=8, step=1, label="Steps")
        guidance = gr.Slider(0.0, 10.0, value=2.5, step=0.1, label="Guidance")
        width = gr.Slider(256, MAX_IMAGE_SIDE, value=768, step=64, label="Width")
        height = gr.Slider(256, MAX_IMAGE_SIDE, value=768, step=64, label="Height")

    seed = gr.Number(value=-1, label="Seed (-1 for random)", precision=0)

    generate_btn = gr.Button("Generate")
    output_image = gr.Image(label="Result")
    output_path = gr.Textbox(label="Saved path", interactive=False)

    generate_btn.click(
        generate_image,
        inputs=[prompt, negative_prompt, steps, guidance, width, height, seed],
        outputs=[output_image, output_path],
    )


demo.launch(
    server_name=os.getenv("GRADIO_SERVER_NAME", "0.0.0.0"),
    server_port=int(os.getenv("GRADIO_SERVER_PORT", "7860")),
)
