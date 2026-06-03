FROM --platform=$BUILDPLATFORM golang:1.23-alpine@sha256:383395b794dffa5b53012a212365d40c8e37109a626ca30d6151c8348d380b5f AS builder

WORKDIR /work
ENV CGO_ENABLED=0

RUN apk add --update --no-cache git

COPY ./go.* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go build -o /app/bot -ldflags "-s -w" .

FROM gcr.io/distroless/static-debian12:latest@sha256:9c346e4be81b5ca7ff31a0d89eaeade58b0f95cfd3baed1f36083ddb47ca3160 AS runtime
WORKDIR /app

COPY --from=builder /app/bot ./
ENTRYPOINT ["/app/bot"]
