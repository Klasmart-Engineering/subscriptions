# syntax=docker/dockerfile:1

FROM golang:1.18-alpine
RUN apk add --update curl && \
    rm -rf /var/cache/apk/*
WORKDIR /app

COPY . .

RUN go get -d -v ./...

RUN go build -o ./athena-mock

EXPOSE 8080

CMD [ "./athena-mock" ]
