ARG NODE_VERSION="20.7.0-alpine"
ARG GO_VERSION="1.23.2-alpine"

# build web app static files.
FROM node:${NODE_VERSION} as web-builder
WORKDIR /rat-web

COPY src/web ./

RUN npm install
RUN npm run build

# build rat server binary, embeding web app static files from previous stage.
FROM golang:${GO_VERSION} as server-builder
WORKDIR /rat

# pre-copy/cache go.mod for pre-downloading dependencies and only
# redownloading them in subsequent builds if they change
COPY src/go.mod src/go.sum ./
RUN go mod download && go mod verify

COPY src .
COPY --from=web-builder /rat-web/build ./web/build

ARG RAT_VERSION="v0.0.0+unknown"
RUN go build \
    -ldflags "-X rat/buildinfo.version=$RAT_VERSION" \
    -v \
    -o /rat/rat

ENV API_AUTHORITY=http://localhost:8877
ENV WEB_AUTHORITY=http://localhost:8888

# build final image
FROM scratch
WORKDIR /rat

COPY --from=server-builder /rat/rat rat
# COPY --from=server-builder /rat/config.yaml config.yaml
COPY config-docker.yaml config.yaml

# 8888 web
# 8877 api
EXPOSE 8888 8877

CMD ["./rat", "-c", "./config.yaml"]
