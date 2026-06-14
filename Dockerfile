FROM golang:1.22-alpine AS builder
RUN apk add --no-cache make git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/lhd ./cmd/lhd

FROM scratch
COPY --from=builder /build/lhd /lhd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/lhd"]
CMD ["--help"]
