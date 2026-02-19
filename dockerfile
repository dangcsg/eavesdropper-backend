# ---------- Build stage ----------
FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build the main in the current module root (adjust if your main is under cmd/)
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server .

# ---------- Runtime stage ----------
FROM debian:12-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /server /server

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/server"]
