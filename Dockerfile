FROM golang:1.25-alpine AS build
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
      -X github.com/miladbeigi/vaultui/internal/version.Version=${VERSION} \
      -X github.com/miladbeigi/vaultui/internal/version.Commit=${COMMIT} \
      -X github.com/miladbeigi/vaultui/internal/version.Date=${DATE}" \
    -o /vaultui .

FROM alpine:3.21
COPY --from=build /vaultui /usr/local/bin/vaultui
ENTRYPOINT ["vaultui"]
