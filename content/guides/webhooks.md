# Qaragon tenant webhooks

Use `list_webhook_events` to discover supported event types and their summaries. They are public subscription types, not internal Kafka topics. Subscribe only to those public names.

Qaragon signs each delivery with `X-FC-Signature` in this format:

```text
t=<unix-seconds>,v1=<hex HMAC-SHA256 of "<t>.<raw-body>">
```

The HMAC key is the subscription signing secret (`whsec_…`). Verify the timestamp tolerance and HMAC in constant time using the exact received bytes, before decoding the body. Keep the signing secret in the tenant's secret manager. Qaragon currently supports owner-only re-reveal of the existing secret; it does not expose a secret-rotation operation. Return a successful response promptly and process heavier work asynchronously. Deduplicate retries using the stable delivery identifier in the envelope.

Use `describe_webhook_event` for the exact event `data` schema and example. The shared envelope schema is returned with each event; this guide covers signature and delivery behavior.

The Go and TypeScript SDK source includes webhook verification helpers under `tools/sdk-gen/static/` in the platform repository. Use `get_sdk_setup` for current package guidance; package publishing/access is managed separately from this public MCP.
