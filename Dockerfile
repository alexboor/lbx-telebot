FROM golang:1.20-alpine as build

WORKDIR /build
RUN apk add --no-cache ca-certificates git
COPY . .
RUN go mod vendor
RUN go build -mod vendor -ldflags "$LD_FLAGS" -o app cmd/main.go;

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /build/app /telebot
WORKDIR /
ENTRYPOINT ["/telebot"]





