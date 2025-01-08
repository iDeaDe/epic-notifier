FROM golang:1.23-alpine AS build

RUN apk update && apk add git

ADD ../.. /go/src/notifier
WORKDIR /go/src/notifier/app

RUN go mod download

RUN go build -C . -tags timetzdata -o notifier

FROM alpine AS run

WORKDIR /app

ADD https://raw.githubusercontent.com/iDeaDe/nemo.travel.geodata/refs/heads/master/countries.json /app/data/countries.json

COPY --from=build /go/src/notifier/app/template /app/data/template
COPY --from=build /go/src/notifier/app/notifier /app/notifier

ENV WORKDIR=/app/data

ENTRYPOINT ["/app/notifier"]