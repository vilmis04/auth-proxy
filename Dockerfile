# syntax=docker/dockerfile:1.7
FROM golang:1.24-alpine@sha256:8bee1901f1e530bfb4a7850aa7a479d17ae3a18beb6e09064ed54cfd245b7191 AS build
WORKDIR /src

# Dependencies first: this layer only reruns when go.mod or go.sum change.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/app /app
ENV GIN_MODE=release \
    PORT=8080
EXPOSE 8080
USER nonroot:nonroot
# Distroless has no curl or wget, so the binary probes itself.
HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
    CMD ["/app", "-healthcheck"]
ENTRYPOINT ["/app"]
