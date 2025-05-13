FROM golang:1.24 as builder

ENV GOPROXY http://proxy.golang.org

RUN mkdir -p /app

WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

ADD . .

RUN make build

FROM gcr.io/distroless/base

WORKDIR /app

COPY --from=builder /app/bin/ .
COPY --from=builder /app/config.yaml .

ENTRYPOINT ["./meu_pau_no_seu_bot"]
CMD ["-config", "./config.yaml"]
