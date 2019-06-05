.PHONY: deps clean build

deps:
	go get -u ./...

clean: 
	rm -rf ./room-conditions/room-conditions
	
build:
	GOOS=linux GOARCH=amd64 go build -o room-conditions/room-conditions ./room-conditions
