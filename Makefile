build:
	go build -o ./viya4-orders-cli main.go

run:
	#TODO: Add your command and order number plus relevant flags to the run command!
	./viya4-orders-cli

all: build run