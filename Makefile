run:
	go run main.go

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/mq-thinktalk main.go
