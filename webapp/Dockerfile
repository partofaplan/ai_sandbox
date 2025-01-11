FROM python:3.10-slim

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE 1  # Prevents Python from writing .pyc files
ENV PYTHONUNBUFFERED 1        # Ensures stdout and stderr are flushed immediately

# Set the working directory
WORKDIR /app

RUN apt-get update && apt-get -y install curl

# Install dependencies
COPY requirements.txt /app/
RUN pip install --no-cache-dir -r requirements.txt

# Copy the application code
COPY . /app/

# Expose the Flask app's port
EXPOSE 6500

CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:6500", "--forwarded-allow-ips", "*", "--proxy-allow-from", "*", "app:app", "--log-level", "debug", "--access-logfile", "-", "--error-logfile", "-", "--capture-output", "--timeout", "120"]

