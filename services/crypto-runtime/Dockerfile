FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/crypto-runtime .

FROM alpine:3.20
RUN adduser -D -H -u 10001 arena
COPY --from=build /out/crypto-runtime /usr/local/bin/crypto-runtime
USER arena
EXPOSE 9090
ENTRYPOINT ["/usr/local/bin/crypto-runtime"]
