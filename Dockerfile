FROM oven/bun:latest AS frontend

WORKDIR /app/web
COPY web/package.json ./
RUN bun install
COPY web ./
COPY VERSION ../
RUN DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat ../VERSION) bun run build

FROM golang:alpine AS backend

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /app

ADD go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN go build -ldflags "-s -w -X 'tea-api/common.Version=$(cat VERSION)'" -o tea-api

FROM alpine

RUN apk add --no-cache ca-certificates tzdata ffmpeg \
    && addgroup -S tea && adduser -S tea -G tea \
    && mkdir /data && chown tea:tea /data \
    && update-ca-certificates

COPY --from=backend /app/tea-api /usr/local/bin/tea-api
USER tea
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["tea-api"]
