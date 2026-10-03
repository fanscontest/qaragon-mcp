# Tenant BFF integration

Tenant applications call Qaragon from a tenant-owned backend for frontend (BFF). The browser or mobile app calls that BFF; the BFF authenticates its user, checks tenant permissions, and calls Qaragon server-to-server.

## Request path

1. Authenticate the end user using the tenant's own identity system.
2. Authorize the requested action using tenant-owned roles and policy.
3. Call the Qaragon API from the BFF over HTTPS with the tenant's `fcn_*` API key in the `Authorization: Bearer ...` header.
4. When an operation acts for a fan, send `X-Acting-As: <piid>` using the platform identity ID. The tenant BFF must first prove that the authenticated tenant user is allowed to act for that identity.
5. Return only the data the tenant client needs. Never expose the tenant API key to a browser, mobile binary, logs, or generated client code.

The MCP server publishes documentation and public API contracts only. It does not authenticate tenants, access tenant data, or proxy API requests.

## Contract-first development

Use `qaragon://api/openapi.json` as the source of truth for paths, operation IDs, authentication, parameters, and request and response schemas. Use the `search_api` tool to find an operation and `describe_operation` for its full contract. Do not infer endpoints from examples or internal service names.

Qaragon responses use the envelope and pagination shapes declared by each operation's schemas. Check those schemas before mapping responses into tenant-specific types. Treat PIIDs as platform identifiers, not tenant database primary keys.

## Retries and changes

For operations whose contract accepts `Idempotency-Key`, send a stable key for retries. The server's replay window and behavior are documented on that operation; do not assume request bodies are compared unless the contract says so. Prefer update operations for profile mutations when the contract specifies them.

## Webhooks

Subscribe only to event types returned by `list_webhook_events`. Verify every delivery against its exact raw body and `X-FC-Signature` before processing. Use a maintained Qaragon SDK helper when available; do not parse and re-serialize JSON before signature verification. Make webhook processing idempotent because deliveries can be retried.
