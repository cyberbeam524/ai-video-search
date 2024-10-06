# Use Python image as the base
FROM python:3.9-slim

# Set the working directory inside the container
WORKDIR /app

# Copy the requirements file and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the Flask app
COPY . .

# Expose the port the app runs on
EXPOSE 5000

# Start the Flask app
CMD ["python", "feature_search_api.py"]
