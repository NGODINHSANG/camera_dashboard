# Use Alpine and download MediaMTX binary
FROM alpine:latest

# Install wget and ca-certificates
RUN apk add --no-cache wget ca-certificates

# Download MediaMTX binary
ARG MEDIAMTX_VERSION=v1.11.3
RUN wget -qO- "https://github.com/bluenviron/mediamtx/releases/download/${MEDIAMTX_VERSION}/mediamtx_${MEDIAMTX_VERSION}_linux_amd64.tar.gz" \
    | tar xz -C /usr/local/bin mediamtx

# Run MediaMTX
ENTRYPOINT ["/usr/local/bin/mediamtx"]
