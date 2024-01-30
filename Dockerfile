FROM golang:1.21-alpine as build
ARG version=0.0.0

WORKDIR /build
RUN apk add --no-cache ca-certificates git
COPY . .
RUN go mod vendor
# TODO add testing
RUN go build -mod vendor -ldflags "$LD_FLAGS" -o app cmd/main.go;

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /build/app /telebot
RUN echo "$version" >> /telebot.version
WORKDIR /
ENTRYPOINT ["/telebot"]





