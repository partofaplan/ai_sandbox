FROM python:3.11-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY requirements.txt ./
RUN pip install --no-cache-dir --index-url https://download.pytorch.org/whl/cpu torch==2.4.0 \
    && pip install --no-cache-dir -r requirements.txt

COPY app.py ./

EXPOSE 7860

CMD ["python", "app.py"]
