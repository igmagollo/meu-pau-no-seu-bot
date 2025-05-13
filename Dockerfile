FROM golang:1.24 as builder

ENV GOPROXY http://proxy.golang.org

RUN mkdir -p /app

WORKDIR /app

COPY go.mod go.mod

COPY go.sum go.sum

ADD . .

RUN CGO_ENABLED=0 go build -o bin/meu-pau-no-seu-bot cmd/meu_pau_no_seu_bot/main.go

FROM scratch

WORKDIR /app

COPY --from=builder bin/meu-pau-no-seu-bot ./
COPY --from=builder config.yml ./

CMD ["/app/meu-pau-no-seu-bot -config config.yml"]
