# Use the official Golang image to create a build artifact.
FROM golang:1.17 as builder

# Copy local code to the container image.
WORKDIR /app
COPY . .

# Build the command inside the container.
RUN go mod download
RUN go build -o sessionerr

# Use a minimal alpine image for the final image
FROM alpine:3.14

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/sessionerr /sessionerr

# Run the web service on container startup.
CMD ["/sessionerr"]
