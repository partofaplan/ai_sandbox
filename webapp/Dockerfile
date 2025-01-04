# Use an official Python runtime as a base image
FROM python:3.9-slim

# Set environment variables to prevent Python from writing .pyc files and to ensure output is flushed
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Set the working directory in the container
WORKDIR /app

# Copy the application code to the container
COPY . /app

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Expose the port Flask will run on
EXPOSE 6500

# Run the Flask application
CMD ["python", "app.py"]