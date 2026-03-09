FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /lokai ./cmd/lokai

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

COPY --from=builder /lokai /usr/local/bin/lokai

# Default Ollama host for container networking
ENV OLLAMA_HOST=http://ollama:11434

ENTRYPOINT ["lokai"]
