# ---- Builder Stage ----
FROM golang:1.22-alpine AS builder

# Install git for private module cloning if needed during build, and for the app itself
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for static linking, making the binary portable
# -ldflags="-s -w" strips debugging information, reducing binary size
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o /social_poster main.go

# ---- Final Stage ----
FROM alpine:latest

# Install git as it's a runtime dependency for git operations
RUN apk add --no-cache git

# Add any other necessary runtime dependencies here
# For example, if you were dealing with images, you might need image processing libraries
# RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /social_poster /app/social_poster

# Copy .env.example for reference (optional)
# COPY .env.example .env.example

# Set the default command for the container
# The application will read its config from an .env file (mounted or present) or environment variables
ENTRYPOINT ["/app/social_poster"]

# Default to running the process command. Users can override this.
CMD ["process"]
