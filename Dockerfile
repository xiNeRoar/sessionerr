# Use the official Golang image to create a build artifact.
FROM golang:1.19 as builder

# Copy local code to the container image.
WORKDIR /app
COPY . .

# Build the command inside the container.
RUN go mod download
RUN go build -o sessionerr

# Use a Debian image for the final image as it has cron available
FROM debian:buster-slim

# Install cron
RUN apt-get update && apt-get install -y cron && rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/sessionerr /sessionerr

# Copy the crontab file
COPY crontab /etc/cron.d/sessionerr-cron
