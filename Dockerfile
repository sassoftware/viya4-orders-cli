FROM golang:1.13 AS builder

WORKDIR /gobuild
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-w -s" -o /viya4-orders-cli

FROM scratch
COPY --from=builder /viya4-orders-cli /usr/bin/viya4-orders-cli
#TODO: Add your command and order number plus relevant flags to the command!
CMD ["/usr/bin/viya4-orders-cli"]