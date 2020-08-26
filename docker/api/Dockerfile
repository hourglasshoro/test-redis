FROM golang:1.14-alpine as build

WORKDIR /go/test-redis

COPY src/app ./src/app
COPY go.mod .
COPY go.sum .
COPY .realize.yml .

RUN apk add --no-cache git \
  && go get github.com/oxequa/realize

WORKDIR /go/test-redis/src/app

RUN go build -o app

FROM alpine as prod

WORKDIR /app

COPY --from=build /go/test-redis/app .

RUN addgroup go \
  && adduser -D -G go go \
  && chown -R go:go /test-redis/app

CMD ["./app"]