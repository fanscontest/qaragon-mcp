# Qaragon public MCP

Connect an MCP client to the Qaragon public developer server over Streamable HTTP. The local development endpoint is `http://localhost:8085/mcp`; hosted environments use the HTTPS hostname configured by Qaragon.

The endpoint is public and requires no developer account or MCP credential. It publishes Qaragon's public API contract and developer guidance. It does not call the API, access tenant data, or accept tenant keys.

## Available tools

- `search_api` finds public operations by feature, method, path, operation ID, or tag.
- `describe_operation` returns one operation's exact request/response details, security requirements, and referenced schemas.
- `get_sdk_setup` provides Go or TypeScript SDK installation guidance.
- `list_webhook_events` lists public tenant webhook event names and descriptions.
- `describe_webhook_event` returns a public event's payload schema/example and the shared delivery envelope schema.

## Available resources and prompt

- `qaragon://api/openapi.json` is the full tenant-facing OpenAPI contract.
- `qaragon://guides/tenant-bff`, `qaragon://guides/webhooks`, `qaragon://guides/sdk-go`, and `qaragon://guides/sdk-typescript` provide integration details.
- `qaragon://guides/public-mcp` explains the MCP itself.
- `build_tenant_integration` prepares a tenant BFF design prompt from a language and app goal.

Ask your MCP client to find an operation for a feature, explain its schemas, or outline a tenant BFF flow. Keep user authentication and tenant policy in your own BFF, and keep the tenant API key server-side.
