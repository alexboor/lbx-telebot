FROM golang:1.21-alpine as build
ARG VERSION=0.0.0

WORKDIR /build
RUN apk add --no-cache ca-certificates git
COPY . .
RUN go mod vendor
# TODO add testing
RUN go build -mod vendor -ldflags "$LD_FLAGS" -o app cmd/main.go;
RUN echo "$VERSION" >> app.version

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /build/app /app
COPY --from=build /build/app.version /app.version
WORKDIR /
ENTRYPOINT ["/app"]





