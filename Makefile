VERSION=$(shell git describe --tags --always)

build:
	mkdir -p bin/ && CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

build-docker:
	docker build -t igmagollo/meu-pau-no-seu-bot:$(VERSION) .

clean:
	rm -f bin/meu-pau-no-seu-bot
