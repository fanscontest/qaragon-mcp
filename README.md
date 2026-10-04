# Qaragon public MCP

Stateless MCP server for tenant developers building applications against Qaragon's public platform APIs. It serves a merged OpenAPI bundle and reviewed integration guidance over Streamable HTTP.

The server is documentation-only. It has no developer login, stores no credentials, and makes no Qaragon API calls. Tenant applications continue to call Qaragon from their own BFFs with their own tenant API key and authorization checks.

## Local development

This service is a separate platform-owned repository checked out as `infra/mcp`, alongside the `uman`, `puzzle`, and `auth` source repositories.

```sh
cd ../infra
task mcp:contract
docker compose up --build qaragon-mcp
```

Connect an MCP client to `http://localhost:8085/mcp`. The health endpoint is `http://localhost:8085/healthz`, and the public OpenAPI document is `http://localhost:8085/openapi.json`. A full local stack starts the service with `task up` after generating both public catalog inputs.

The service expects these files at startup:

- `generated/public-openapi.json`: the merged, public-auth-overlaid contract from `tools/sdk-gen/assemble-live.sh`.
- `generated/webhook-events.json`: public webhook event names extracted from uman's tenant webhook allow-list.

## Interface

Tools: `search_api`, `describe_operation`, `get_sdk_setup`, and `list_webhook_events`.

Resources: `qaragon://api/openapi.json` and the curated `qaragon://guides/*` documents.

Prompt: `build_tenant_integration`.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | HTTP listener |
| `OPENAPI_PATH` | `/app/generated/public-openapi.json` | Public contract input |
| `WEBHOOK_EVENTS_PATH` | `/app/generated/webhook-events.json` | Public webhook event list |
| `GUIDES_DIR` | `/app/content/guides` | Curated Markdown bundle |
| `PLATFORM_API_URL` | `localhost` | Public API hostname in the OpenAPI `servers` field; use a bare hostname |
| `MCP_ALLOWED_HOSTS` | `localhost,127.0.0.1,mcp.qaragon.com` | Host header allow-list |

Staging exposes the API and OpenAPI document at `https://api.staging.qaragon.com`
and MCP at `https://mcp.staging.qaragon.com/mcp`. Production uses
`https://api.qaragon.com` and `https://mcp.qaragon.com/mcp`.

The edge overwrites `X-Real-IP` before forwarding requests. Request bodies are capped at 1 MiB, and the process applies an in-memory per-IP request limit. No requests or tool inputs are persisted.
