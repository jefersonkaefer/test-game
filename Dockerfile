FROM golang:1.23

WORKDIR /go/src/game

COPY backend/app/ .

RUN ls -la && ls -la cmd/api/

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/main ./cmd/api/main.go
RUN ls -la /go/bin/

FROM alpine:3.14
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=0 /go/bin/main .

EXPOSE 8000

CMD ["./main"]