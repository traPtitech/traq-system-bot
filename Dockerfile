FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

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

FROM gcr.io/distroless/static-debian12:latest AS runtime
WORKDIR /app

COPY --from=builder /app/bot ./
ENTRYPOINT ["/app/bot"]
