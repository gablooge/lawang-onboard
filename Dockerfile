# Build stage: static Go binary.
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/onboard ./cmd/onboard

# Final stage: distroless, no shell, runs as nonroot.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/onboard /app/onboard
COPY --chmod=0555 corpus/*.jsonl corpus/manifest.json /app/corpus/
COPY --chmod=0444 roles.yaml /app/roles.yaml
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/app/onboard","-addr",":8080","-corpus","/app/corpus","-roles","/app/roles.yaml"]
