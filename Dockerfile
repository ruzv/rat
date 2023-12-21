# build web app static files.
FROM node:20.7.0-alpine as web-builder
WORKDIR /rat-web

COPY src/web ./

RUN npm install
RUN npm run build

# build rat server binary, embeding web app static files from previous stage.
FROM golang:1.21-alpine as server-builder
WORKDIR /rat

# pre-copy/cache go.mod for pre-downloading dependencies and only
# redownloading them in subsequent builds if they change
COPY src/go.mod src/go.sum ./
RUN go mod download && go mod verify

COPY src .
COPY --from=web-builder /rat-web/build ./web/build

ARG RAT_VERSION="v0.0.0+unknown"
RUN go build -ldflags "-X main.version=$RAT_VERSION" -v -o /rat/rat

# build final image
FROM scratch
WORKDIR /rat

COPY --from=server-builder /rat/rat rat
COPY config-docker.yaml config.yaml

EXPOSE 8888

CMD ["./rat", "-c", "./config.yaml"]
