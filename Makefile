
BINARY_NAME := rat
VERSION := $(shell git describe --tags)

DOCKER_IMAGE_NAME := ruzv/rat
DOCKER_IMAGE_TAG := dev
DOCKER_IMAGE_NODE_VERSION := 20.7.0-alpine
DOCKER_IMAGE_GO_VERSION := 1.23.2-alpine

build-image:
	docker build \
		--build-arg RAT_VERSION=${VERSION} \
		--build-arg NODE_VERSION=${DOCKER_IMAGE_NODE_VERSION} \
		--build-arg GO_VERSION=${DOCKER_IMAGE_GO_VERSION} \
		-t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .
.PHONY: build-image

build-binary: build-binary-web build-binary-server
.PHONY: build-binary

build-binary-web:
	npm --prefix src/web install
	npm --prefix src/web run build
.PHONY: build-binary-web

build-binary-server:
	cd src && go build \
		-v \
		-ldflags "-X rat/buildinfo.version=${VERSION}" \
		-o ${BINARY_NAME}
	mv src/${BINARY_NAME} .
.PHONY: build-binary-server
