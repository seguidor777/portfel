IMAGE   ?= ghcr.io/seguidor777/portfel
TAG     ?= latest
PLATFORM ?= linux/amd64

.PHONY: build push pull up down logs test

## Build the container image
build:
	docker build --platform $(PLATFORM) -t $(IMAGE):$(TAG) .

## Push the image to the registry
push:
	docker push $(IMAGE):$(TAG)

## Pull the latest image from the registry (use on the VM instead of building)
pull:
	docker pull $(IMAGE):$(TAG)

## Start the bot (detached)
up:
	nerdctl compose up -d

## Stop the bot
down:
	nerdctl compose down

## Tail live logs
logs:
	nerdctl compose logs -f portfel

## Run the full test suite (outside Docker)
test:
	go test ./...
