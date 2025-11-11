FROM golang:1.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/mcp-templates ./cmd

FROM alpine:3.19

RUN addgroup -S app && adduser -S -G app app && apk add --no-cache ca-certificates

WORKDIR /workspace

COPY --from=builder /out/mcp-templates /usr/local/bin/mcp-templates
COPY templates ./templates

RUN chown -R app:app /workspace

USER app

ENV TEMPLATES_PATH=/workspace/templates \
    OBS_ENABLE_METRICS=true \
    OBS_METRICS_ADDRESS=:2112 \
    OBS_SERVICE_NAME=mcp-template-cli \
    OBS_ENABLE_TRACING=false \
    OBS_OTLP_ENDPOINT=jaeger:4317

EXPOSE 2112

ENTRYPOINT ["/usr/local/bin/mcp-templates"]
CMD ["--help"]

