# Use the official Go image as a base.
FROM golang:1.21-alpine

# Set the working directory inside the container.
WORKDIR /app

# Install 'air', a live-reloading tool for Go applications.
RUN go install github.com/cosmtrek/air@latest

# Copy the go.mod and go.sum files to cache dependencies.
COPY ./client/go.mod ./client/go.sum ./
RUN go mod download

# Copy the rest of your client source code.
COPY ./client ./

# Copy the air.toml file.
COPY ./client/deployment/.air.toml ./.air.toml

# The entrypoint will be 'air', which watches for file changes.
# It will use the .air.toml configuration file.
ENTRYPOINT ["air", "-c", ".air.toml"]
