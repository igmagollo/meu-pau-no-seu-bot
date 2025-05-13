VERSION=$(shell git describe --tags --always)

build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...	

clean:
	rm -f bin/meu-pau-no-seu-bot
