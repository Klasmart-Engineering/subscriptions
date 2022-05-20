# syntax=docker/dockerfile:1

FROM golang:1.18-alpine
RUN apk add --update curl && \
    rm -rf /var/cache/apk/*
WORKDIR /app

COPY ./src .
COPY environment/local/.env .

RUN go get -d -v ./...

RUN go install -v ./...

RUN go build -o /subscriptions-app


EXPOSE 8080

CMD [ "/subscriptions-app" ]
