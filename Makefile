GIT_COMMIT_COUNT=$(shell printf '%05d' `git rev-list --count HEAD`)
GIT_COMMIT_TIME=$(shell git show -s --date=format:'%Y%m%d-%H%M' --format='%cd')
GIT_SHA=$(shell git rev-parse --short HEAD)

IMAGE_TAG?=${GIT_COMMIT_COUNT}-${GIT_SHA}
IMAGE_REPOSITORY?=ghcr.io/akuity/guestbook

.PHONY: build
build:
	CGO_ENABLED=0 go build -v -o build/guestbook ./main.go

.PHONY: image
image: build
	docker build -t ${IMAGE_REPOSITORY}:${IMAGE_TAG} .

# push latest image with multiple tags
# only push on linux/amd64 so we don't need to deal with multi-arch images
.PHONY: push-latest
push-latest:
	docker buildx build --push --platform=linux/amd64 . \
		-t ${IMAGE_REPOSITORY}:${GIT_COMMIT_COUNT}-${GIT_SHA} \
		-t ${IMAGE_REPOSITORY}:${GIT_COMMIT_COUNT} \
		-t ${IMAGE_REPOSITORY}:${GIT_COMMIT_TIME}-${GIT_SHA} \
		-t ${IMAGE_REPOSITORY}:latest
