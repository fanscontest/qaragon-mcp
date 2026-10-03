FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/qaragon-mcp ./cmd/qaragon-mcp

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget \
    && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/qaragon-mcp /app/qaragon-mcp
COPY generated/ /app/generated/
COPY content/guides/ /app/content/guides/
USER app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/qaragon-mcp"]
