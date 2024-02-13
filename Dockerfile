FROM golang:1.22.0-alpine as build

ADD . /go/src/notifier
WORKDIR /go/src/notifier

RUN apk update && \
    apk add git && \
    go get && \
    go build -o notifier

FROM alpine as run

WORKDIR /app
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /app/zoneinfo.zip
COPY --from=build /go/src/notifier/template /app/data/template
COPY --from=build /go/src/notifier/notifier /app/notifier

ENV ZONEINFO=/app/zoneinfo.zip
ENV WORKDIR=/app/data

VOLUME ["/app/data/"]
ENTRYPOINT ["./notifier"]