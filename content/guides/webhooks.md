# Qaragon tenant webhooks

Use `list_webhook_events` to read the public event allow-list extracted from Qaragon's tenant webhook policy. The names exposed there are subscription event types, not internal Kafka topics. Subscribe only to those public names.

Qaragon signs each delivery with `X-FC-Signature` in this format:

```text
t=<unix-seconds>,v1=<hex HMAC-SHA256 of "<t>.<raw-body>">
```

The HMAC key is the subscription signing secret (`whsec_…`). Verify the timestamp tolerance and HMAC in constant time using the exact received bytes, before decoding the body. Keep the signing secret in the tenant's secret manager and support secret rotation. Return a successful response promptly and process heavier work asynchronously. Deduplicate retries using the stable event identifier in the payload.

The Go and TypeScript SDK source includes webhook verification helpers under `tools/sdk-gen/static/` in the platform repository. Use `get_sdk_setup` for current package guidance; package publishing/access is managed separately from this public MCP.
