# Build stage: static Go binary, cross-compiled so one build serves amd64 and arm64.
FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-s -w" -o /out/onboard ./cmd/onboard

# Final stage: distroless, no shell, runs as nonroot.
FROM gcr.io/distroless/static-debian12:nonroot
LABEL io.modelcontextprotocol.server.name="io.github.gablooge/lawang-onboard" \
      org.opencontainers.image.source="https://github.com/gablooge/lawang-onboard" \
      org.opencontainers.image.description="Permission-aware onboarding MCP server: answers about a codebase, filtered by the caller's role." \
      org.opencontainers.image.licenses="MIT"
WORKDIR /app
COPY --from=build /out/onboard /app/onboard
COPY --chmod=0555 corpus/*.jsonl corpus/manifest.json /app/corpus/
COPY --chmod=0444 roles.yaml /app/roles.yaml
USER nonroot
EXPOSE 47312
ENTRYPOINT ["/app/onboard","-addr",":47312","-corpus","/app/corpus","-roles","/app/roles.yaml"]
