FROM golang:1.20 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build

FROM scratch
COPY --from=builder /app/att-fiber-prom-exporter /att-fiber-prom-exporter
ENTRYPOINT ["/att-fiber-prom-exporter"]
