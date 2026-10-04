FROM node:22-alpine AS swagger-ui

WORKDIR /tmp/swagger-ui
RUN npm pack swagger-ui-dist@5.33.1 --pack-destination /tmp \
    && mkdir -p /out \
    && tar -xzf /tmp/swagger-ui-dist-5.33.1.tgz --strip-components=1 -C /out

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
RUN mkdir -p /app/content/swagger-ui
COPY --from=swagger-ui /out/swagger-ui.css /app/content/swagger-ui/
COPY --from=swagger-ui /out/swagger-ui-bundle.js /app/content/swagger-ui/
COPY --from=swagger-ui /out/swagger-ui-standalone-preset.js /app/content/swagger-ui/
COPY generated/ /app/generated/
COPY content/guides/ /app/content/guides/
USER app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/qaragon-mcp"]
