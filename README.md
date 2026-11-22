# Qaragon public MCP

Stateless MCP server for tenant developers building applications against Qaragon's public platform APIs. It serves a merged OpenAPI bundle and reviewed integration guidance over Streamable HTTP.

The server is documentation-only. It has no developer login, stores no credentials, and makes no Qaragon API calls. Tenant applications continue to call Qaragon from their own BFFs with their own tenant API key and authorization checks.

The [Qaragon Developer plugin](docs/developer-plugin.md) bundles this connection with a tenant-integration skill and examples. It can be installed from this checkout's marketplace for local use in a coding assistant.

## Local development

This service is a separate platform-owned repository checked out as `infra/mcp`, alongside the `core`, `puzzle`, and `auth` source repositories.

```sh
cd ../infra
task mcp:contract
docker compose up --build qaragon-mcp
```

Connect an MCP client to `http://localhost:8085/mcp`. The service also hosts an interactive API reference at `http://localhost:8085/`, the machine-readable OpenAPI contract at `http://localhost:8085/openapi.json`, and a health endpoint at `http://localhost:8085/healthz`. A full local stack starts the service with `task up` after generating both public catalog inputs.

The service expects these files at startup:

- `generated/public-openapi.json`: the merged, public-auth-overlaid contract from `tools/sdk-gen/assemble-live.sh`.
- `generated/webhook-events.json`: public webhook event schemas/examples validated against core's allow-list and payload encoders.

## Interface

Tools: `search_api`, `describe_operation`, `get_sdk_setup`, `list_webhook_events`, and `describe_webhook_event`.

Resources: `qaragon://api/openapi.json` and the curated `qaragon://guides/*` documents, including `qaragon://guides/find-the-minority`, `qaragon://guides/fans-say` `qaragon://guides/fans-feud` and `qaragon://guides/word-chain` and `qaragon://guides/moment`.

Prompt: `build_tenant_integration`.

Qaragon's domain glossary and Contest rules are published as `qaragon://guides/domain-language` and `qaragon://guides/contest-rules`. Every tenant-integration prompt directs the agent to the relevant glossary sections and includes the Contest rules. Their canonical sources are `core/docs/domain-language.md` and `core/docs/contest-rules.md`. From the infra checkout, run `task mcp:guides` after updating those sources; `task mcp:contract` also syncs them automatically. Edit the canonical sources rather than the bundled copies. The developer plugin reads the same MCP resources.

The Find the Minority workflow is published as `qaragon://guides/find-the-minority`. Its canonical source is `core/docs/games/find-the-minority.md`; copy it into `content/guides/find-the-minority.md` when updating the tenant-facing flow. The public OpenAPI bundle is assembled separately from the live service contracts.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | HTTP listener |
| `OPENAPI_PATH` | `/app/generated/public-openapi.json` | Public contract input |
| `WEBHOOK_EVENTS_PATH` | `/app/generated/webhook-events.json` | Public webhook event list |
| `GUIDES_DIR` | `/app/content/guides` | Curated Markdown bundle |
| `PLATFORM_API_URL` | `localhost` | Public API hostname in the OpenAPI `servers` field; use a bare hostname |
| `MCP_ALLOWED_HOSTS` | `localhost,127.0.0.1,mcp.qaragon.com` | Host header allow-list |

Staging exposes the interactive API reference at `https://api.staging.qaragon.com/`,
the OpenAPI document at `https://api.staging.qaragon.com/openapi.json`, and MCP at
`https://mcp.staging.qaragon.com/mcp`. Production uses
`https://api.qaragon.com` and `https://mcp.qaragon.com/mcp`.

The edge overwrites `X-Real-IP` before forwarding requests. Request bodies are capped at 1 MiB, and the process applies an in-memory per-IP request limit. No requests or tool inputs are persisted.

The word-game guides cover full and standalone offerings: Fans Say, Seven Clues, Ten in Sixty, Fans Feud, Feud Showdown, Feud Rush, Link Up, Link Clash and Link Rush. Their canonical sources are `core/docs/games/fans-say.md`, `core/docs/games/fans-feud.md` and `core/docs/games/link-up.md`. Keep the copies in `content/guides/` aligned when updating the API flow. The integration prompt includes the matching Game guide for these names.

### Game guide rollout order

Deploy Core with Fans Feud support (PR #325) before releasing the Fans Feud MCP guide. Merging the backend PR alone does not make its Game types available on the API. Verify the target environment exposes those types before publishing the matching guide and integration prompt.

Community Word Reveal uses `kind=word_reveal` and the `qaragon://guides/community-guess` resource. Fans Say (`type=fans_say`) remains a hosted Game with its own guide. Integration prompts select the community guide for “Word Reveal” or `word_reveal` requests.

Shared Word Chain and Word Reveal publication approval is documented at
`qaragon://guides/content-review`. Qaragon platform administrators review generated
candidates; tenant app-admin status grants no approval rights. Keep the canonical
`core/docs/games/content-review.md` synchronized with `task mcp:guides`.

The Automated Contest Scheduler, its settings, dated reads and bounded replenishment
are documented in `qaragon://guides/recurring-contests`, synchronized from the
canonical Core guide by `task mcp:guides`. Daily/recurring and scheduler-settings integration prompts
include this guide. Global paid generation is disabled unless explicitly configured. Tenants opt in
separately; FansContest creates its Nigeria preferences once through its own
bootstrap.
