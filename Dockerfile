FROM golang:1.25-alpine AS build
RUN apk add --no-cache git && CGO_ENABLED=0 go install github.com/miladbeigi/vaultui@latest

FROM alpine:3.21
COPY --from=build /go/bin/vaultui /usr/local/bin/vaultui
ENTRYPOINT ["vaultui"]
