# build web app static files.
FROM node:20.7.0-alpine as web-builder

WORKDIR /usr/src/rat-web

COPY src/web ./

RUN npm install
RUN npm run build

# build rat server binary, embeding web app static files from previous stage.
FROM golang:1.21-alpine as server-builder

WORKDIR /usr/src/rat

# pre-copy/cache go.mod for pre-downloading dependencies and only
# redownloading them in subsequent builds if they change
COPY src/go.mod src/go.sum ./
RUN go mod download && go mod verify

COPY src .

COPY --from=web-builder /usr/src/rat-web/build ./web/build

RUN go build -v -o /usr/local/bin/rat

# build final image
FROM alpine:3.19.0

COPY --from=server-builder /usr/local/bin/rat /usr/local/bin/rat

COPY docker/config.yaml /etc/rat-config.yaml

EXPOSE 8888

CMD ["/usr/local/bin/rat", "-c", "/etc/rat-config.yaml"]

# docker run -it --entrypoint sh rat
