---
name: build-qaragon-integration
description: Build or troubleshoot a tenant application's Qaragon integration, including its BFF, contest or hosted-game flow, submissions, results and webhooks. Use when integrating an app with Qaragon's public API.
---

# Build with Qaragon

Use the tenant's existing application and BFF. The `qaragon` MCP server supplies public documentation; implementation happens in the tenant's repository. Platform credentials belong in the tenant's server-side configuration, separate from this credential-free MCP connection.

## Establish the flow

Inspect the existing auth, SDK/transport, identity mapping and UI before choosing an implementation. Determine the feature and target API environment from the user's request and project configuration. Ask only about decisions that change the implementation and remain unclear.

Distinguish an ordinary Contest from a hosted Game. Community Word Chain, Word Reveal and Fans Feud use ordinary Guess Contests; their hosted counterparts use the Game API. For an ambiguous request, establish the play style before choosing routes. A playable Contest does not need a Game wrapper.

Read the relevant [domain glossary](qaragon://guides/domain-language) sections to establish the capability's terms, ownership and relationships. For Contest configuration, grading, zones, ranking, grouping or winners, also read [Contest rules](qaragon://guides/contest-rules). Establish the intended rule and its scope, then check the type's implementation support. Treat an ignored setting as a support gap to resolve, rather than a reason to redefine the mechanic.

## Consult the current contract

Read [tenant BFF guidance](qaragon://guides/tenant-bff). Call `get_sdk_setup` with `language: "go"` or `"typescript"` when using that SDK. Use `search_api` for the capabilities in this flow and `describe_operation` with each returned `operation_id` before implementing its request or response mapping. The MCP connection's hostname does not choose the tenant application's API environment.

Select the relevant guide through the server's resource list. Community word boards use [community Guess](qaragon://guides/community-guess); Moment uses [Moment](qaragon://guides/moment). Hosted Games have their own guides. Use [integration recipes](references/integration-recipes.md) for community submissions or webhooks, and the [Go SDK examples](references/go-sdk-examples.md) when bootstrapping a Go BFF.

Keep schemas and operation details in MCP rather than copying a contract into the application or this skill. If the MCP is unavailable, use a verified public contract matching the target deployment; otherwise report what cannot yet be verified. When an installed SDK cannot represent the contract, upgrade or regenerate it through the project's SDK workflow and report the gap.

## Implement across the boundary

- The tenant authenticates its users and authorizes actions in its BFF. Resolve the authorized user's PlatformIdentity from the tenant's mapping and pass `X-Acting-As` for fan-scoped calls. A client-supplied PIID alone is not authorization.
- Spaces belong to Venues. Joining a Space and enrolling in a Contest are distinct operations; follow the selected Contest's requirements. Approved membership is required for submissions from that Space.
- Every participant action uses the shared `POST /v2/submissions`, targeting a Contest or Game. Keep the module's declared execution mode, repeatability and timing; a 202 acceptance is not a final result. Preserve the same idempotency key when retrying the same action.
- Follow the operation's result visibility and pending/error responses. Notifications, reward delivery and fan-facing realtime presentation remain tenant-owned integrations.
- For webhooks, discover public names with `list_webhook_events`, inspect payloads with `describe_webhook_event`, then read [webhook delivery guidance](qaragon://guides/webhooks). Verify the exact raw bytes with a current SDK helper and persist the contract's batch/delivery identifiers for retry deduplication.

## Verify the integration

Exercise the actual BFF and app flow: authenticated identity selection, membership/enrollment when required, shared submission acceptance or immediate outcome, pending results and finalized results. For a webhook receiver, test a valid signature, a modified body, a stale timestamp and a repeated delivery. Use the tenant's existing checks and test conventions.

Report changed files, verified operations and SDK version, checks run, and any deployment or credential setup still required. Keep live data writes and deployments within the user's authorization; the documentation MCP performs neither.
