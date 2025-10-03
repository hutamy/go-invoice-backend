FROM golang:1.24.7-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o go-invoice ./cmd/main.go

# Run stage
FROM debian:bookworm-slim

WORKDIR /app

# Install Chrome and dependencies
RUN apt-get update && apt-get install -y \
  chromium \
  ca-certificates \
  fonts-liberation \
  libasound2 \
  libatk-bridge2.0-0 \
  libdrm2 \
  libxcomposite1 \
  libxdamage1 \
  libxrandr2 \
  libgbm1 \
  libxss1 \
  libgtk-3-0 \
  && rm -rf /var/lib/apt/lists/*

# Set Chrome path for chromedp
ENV CHROME_BIN=/usr/bin/chromium

COPY --from=builder /app/go-invoice .
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./go-invoice"]