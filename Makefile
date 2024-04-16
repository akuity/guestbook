IMAGE_TAG?=latest
IMAGE_REPOSITORY?=guestbook

.PHONY: build
build:
	CGO_ENABLED=0 go build -v -o build/guestbook ./main.go

.PHONY: image
image: build
	docker build -t ${IMAGE_REPOSITORY}:${IMAGE_TAG} .