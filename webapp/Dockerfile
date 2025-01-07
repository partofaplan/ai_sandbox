# Use an official Python runtime as the base image
FROM python:3.10-slim

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE 1  # Prevents Python from writing .pyc files
ENV PYTHONUNBUFFERED 1        # Ensures stdout and stderr are flushed immediately

# Set the working directory
WORKDIR /app

# Install dependencies
COPY requirements.txt /app/
RUN pip install --no-cache-dir -r requirements.txt

# Copy the application code
COPY . /app/

# Expose the Flask app's port
EXPOSE 6500

# Set the command to run the app with Gunicorn
CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:6500", "app:app"]
