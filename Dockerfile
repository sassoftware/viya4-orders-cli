# Copyright © 2020, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
FROM golang:1.16 AS builder

WORKDIR /gobuild
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Inject the current version so -v / --version will work.
RUN go build -ldflags "-w -s -X github.com/sassoftware/viya4-orders-cli/cmd.version=$(git describe --tags --abbrev=0 --always)" -o /viya4-orders-cli

# Install certs.
FROM alpine:latest AS certAdder
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# Don't be root!
RUN addgroup -g 1000 -S appuser && \
    adduser -S -u 1000 -G appuser appuser
USER appuser

FROM scratch

# Copy viya4-orders-cli binary in.
COPY --from=builder /viya4-orders-cli /usr/bin/viya4-orders-cli

# Copy certs that we installed earlier in.
COPY --from=certAdder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/usr/bin/viya4-orders-cli"]
CMD ["--help"]
