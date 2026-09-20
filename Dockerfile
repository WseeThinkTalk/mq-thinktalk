FROM golang:1.26-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /app/mq-thinktalk main.go

FROM alpine:3.20

ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app
COPY --from=builder /app/mq-thinktalk /app/mq-thinktalk
COPY etc /app/etc

CMD ["./mq-thinktalk", "-f", "etc/mq.yaml"]
