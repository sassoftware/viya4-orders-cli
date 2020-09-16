build:
	go build -o ./bin/viya4-orders-cli main.go

run:
	./bin/viya4-orders-cli

all: build run