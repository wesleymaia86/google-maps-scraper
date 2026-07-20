# Build stage for Playwright dependencies
FROM ubuntu:20.04 AS playwright-deps
ENV PLAYWRIGHT_BROWSERS_PATH=/opt/browsers
ENV PLAYWRIGHT_DRIVER_PATH=/opt/ms-playwright-go
ARG TARGETARCH
ARG PLAYWRIGHT_GO_VERSION=v0.6100.0

RUN export PATH=$PATH:/usr/local/go/bin:/root/go/bin \
    && apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl wget \
    && if [ "$TARGETARCH" = "arm64" ]; then \
         GO_ARCH="arm64"; \
       else \
         GO_ARCH="amd64"; \
       fi \
    && wget -q "https://go.dev/dl/go1.26.5.linux-${GO_ARCH}.tar.gz" \
    && tar -C /usr/local -xzf "go1.26.5.linux-${GO_ARCH}.tar.gz" \
    && rm "go1.26.5.linux-${GO_ARCH}.tar.gz" \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && go install github.com/mxschmitt/playwright-go/cmd/playwright@${PLAYWRIGHT_GO_VERSION} \
    && mkdir -p /opt/browsers \
    && playwright install chromium --with-deps \
    # The scraping runtime pulls in playwright-community/playwright-go v0.6000.0
    # transitively (via scrapemate), which requires Playwright driver v1.60.0.
    # That version's legacy driver download (playwright.azureedge.net) now 404s,
    # so every browser scrape fails with:
    #   "could not install driver: driver exists but version not 1.60.0".
    # The community fork cannot be bumped to a fixed release either: its v0.61xx
    # tags redeclare the module path as github.com/mxschmitt/playwright-go, so Go
    # rejects them under the community import path. Until scrapemate migrates,
    # supply the exact driver the runtime validates against from npm (still
    # served), then fetch the Chromium build v1.60.0 pins. The runtime only
    # checks `node package/cli.js --version == 1.60.0` before launching.
    && curl -fsSL https://registry.npmjs.org/playwright-core/-/playwright-core-1.60.0.tgz -o /tmp/pw-core.tgz \
    && rm -rf /opt/ms-playwright-go/package \
    && tar -xzf /tmp/pw-core.tgz -C /opt/ms-playwright-go \
    && rm /tmp/pw-core.tgz \
    && /opt/ms-playwright-go/node /opt/ms-playwright-go/package/cli.js --version | grep -q 1.60.0 \
    && /opt/ms-playwright-go/node /opt/ms-playwright-go/package/cli.js install chromium

# Build stage
FROM golang:1.26.5-trixie AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /usr/bin/google-maps-scraper

# Final stage
FROM debian:trixie-slim
ENV PLAYWRIGHT_BROWSERS_PATH=/opt/browsers
ENV PLAYWRIGHT_DRIVER_PATH=/opt/ms-playwright-go

# Install only the necessary dependencies in a single layer
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libnss3 \
    libnspr4 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libdbus-1-3 \
    libxkbcommon0 \
    libatspi2.0-0 \
    libx11-6 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libpango-1.0-0 \
    libcairo2 \
    libasound2 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=playwright-deps /opt/browsers /opt/browsers
COPY --from=playwright-deps /opt/ms-playwright-go /opt/ms-playwright-go

RUN chmod -R 755 /opt/browsers \
    && chmod -R 755 /opt/ms-playwright-go

COPY --from=builder /usr/bin/google-maps-scraper /usr/bin/

ENTRYPOINT ["google-maps-scraper"]
