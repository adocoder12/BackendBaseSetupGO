## The base image: a small Linux (Alpine) with Go already installed
FROM golang:1.27-alpine

## Runs a command while the image is being built. It downloads and installs Air (pinned to one version, so builds are repeatable).
RUN go install github.com/air-verse/air@v1.61.7

## Sets the working folder inside the container. All later commands run from /app, and it's created if missing. This is why compose mounts your project to /app.
WORKDIR /app

## Copies only the dependency files and downloads the modules. This is done before copying your code on purpose, for caching.
COPY go.mod go.sum ./
RUN go mod download

## Copies the rest of your project into /app.
COPY . .

##we The command that runs when the container starts (not at build time)
CMD ["air"]
