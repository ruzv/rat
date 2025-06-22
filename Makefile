
BINARY_DIR := bin
BINARY_NAME := rat
VERSION := $(shell git describe --tags)

DOCKER_IMAGE_NAME := ruzv/rat
DOCKER_IMAGE_TAG := dev
DOCKER_IMAGE_NODE_VERSION := 20.7.0-alpine
DOCKER_IMAGE_GO_VERSION := 1.23.2-alpine

build-image:
	docker build \
		--progress=plain \
		--build-arg RAT_VERSION=${VERSION} \
		--build-arg NODE_VERSION=${DOCKER_IMAGE_NODE_VERSION} \
		--build-arg GO_VERSION=${DOCKER_IMAGE_GO_VERSION} \
		-t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .
.PHONY: build-image

build-binary: build-binary-web build-binary-server
.PHONY: build-binary

build-binary-web:
	npm --prefix web install
	npm --prefix web run build
	rm -rf cmd/rat/embed/*
	cp -r web/build/* cmd/rat/embed
.PHONY: build-binary-web

build-binary-server:
	mkdir -p ${BINARY_DIR}
	go build \
		-v \
		-ldflags "-X rat/buildinfo.version=${VERSION}" \
		-o ${BINARY_DIR}/${BINARY_NAME} \
		./cmd/rat

.PHONY: build-binary-server
