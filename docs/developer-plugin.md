# Qaragon Developer plugin

Qaragon Developer packages the public documentation MCP connection and one `build-qaragon-integration` skill for tenant developers. It helps implement or troubleshoot tenant BFFs and apps using current API contracts, prepared-content or hosted-game guides, shared submissions, results and webhooks.

The package is in `plugin/`; this repository's marketplace is `.agents/plugins/marketplace.json`. The portable `plugin.json` and `mcp.json` use the Agent Plugins 1.0.0 layout. Installation does not require tenant credentials. The MCP remains documentation-only; the coding assistant uses its normal repository tools for authorized implementation work.

## Install the working checkout

From this repository, using a Codex client with plugin support:

```sh
codex plugin marketplace add "$PWD"
codex plugin add qaragon-developer@qaragon
codex plugin list --marketplace qaragon --json
```

Start a new conversation in the tenant application's repository after installation. In a supported desktop client, the plugin is also available from the Qaragon marketplace in the Plugins directory. An installed plugin is a cached copy: after editing this checkout, refresh/reinstall it with the client's plugin controls and start a new conversation before testing the changes.

For a new local install from the CLI, `codex plugin remove qaragon-developer@qaragon` followed by `codex plugin add qaragon-developer@qaragon` refreshes the cached package. This affects this plugin's installation only.

Once this branch is merged/pushed, developers can add the repository marketplace instead of a local checkout:

```sh
codex plugin marketplace add fanscontest/qaragon-mcp
codex plugin add qaragon-developer@qaragon
```

GitHub access to the repository is separate from MCP access. Public-directory publication in ChatGPT/Codex is a later step; this package has not been submitted or published there.

## Connection and environment

Version 0.1.0 is a development preview and connects to `https://mcp.staging.qaragon.com/mcp` for current public contracts and guidance. That does not configure the tenant application's API base URL: choose its deployment URL and credentials in its own BFF configuration. Use a contract aligned with that target deployment.

Both hosted endpoints accepted MCP initialization during the 2026-11-19 checks, but production lacked `describe_webhook_event`; staging passed all five tools and the community integration flow. Before public release, deploy the current MCP/catalog to production, check its tools and referenced resources, then change `plugin/mcp.json` to `https://mcp.qaragon.com/mcp` and refresh the package.

To test a locally changed MCP server, edit the checkout's `plugin/mcp.json` URL to `http://localhost:8085/mcp`, then refresh the installed package. Restore its hosted endpoint afterward. No tenant API key belongs in `mcp.json` or skill metadata.

## Try it in Branda or another tenant

Invoke `$qaragon-developer:build-qaragon-integration`, or select Qaragon Developer, with a concrete integration request:

- “Add non-hosted Word Chain contests to this app, using our existing Go BFF and authentication.” Expect ordinary Guess Contests, approved Space membership, final-board submissions and results after settlement.
- “Add a host-led Fans Say experience.” Expect the hosted Game guide and Game state/actions, while participant actions still use the shared submission endpoint.
- “Review our result webhook receiver.” Expect public event discovery, signature verification against raw bytes, batch-aware deduplication and tenant-owned notification delivery.
- “Our installed SDK cannot decode this current Contest response.” Expect a verified contract/version comparison and the project's SDK upgrade/generation workflow.
- “Add Word Chain” with no indication of play style. Expect inspection of existing app context and, if still ambiguous, a hosted/community clarification before choosing the flow.
- “Put our tenant key in the mobile app” or “Use the MCP to create a live Contest.” Expect the server-side credential boundary and a clear explanation of which API calls or authorization the requested action requires.

Use a small change in a tenant development branch for the first assisted implementation. Check the produced code against the current MCP contract and run that tenant's tests. Package/MCP checks alone do not establish that model behavior or a tenant's deployed flow is correct.

## Maintainer checks

```sh
go test ./...
```

The package tests verify marketplace resolution, its public credential-free MCP configuration, and that packaged links resolve either inside the package or to resources on the real MCP implementation. Validate the skill's frontmatter and check the package with the target client as well. Recheck SDK examples when upgrading the published SDK; their HTTP and signature behavior can be exercised against fixtures without live tenant writes.

Version 0.1.0 checks completed on 2026-11-19:

- `go test ./...` passed, including the package/reference checks.
- The skill-creator validator accepted the skill.
- Codex CLI 0.160.0 installed the package in an isolated profile. Its app server recognized the `qaragon` MCP connection and the enabled `qaragon-developer:build-qaragon-integration` skill.
- The Go examples compiled against published SDK v0.11.0. Fixture tests with the race detector verified tenant auth, distinct acting identities for concurrent calls, valid signatures, modified bodies and stale timestamps.
- Live read-only staging checks exercised all five MCP tools, the referenced guides, and the non-hosted Word Chain integration prompt. Production still needs the MCP update described above.

These checks did not install the plugin into the user's main profile, call tenant mutation endpoints, run a model-assisted Branda implementation, or publish the package.

Keep skill instructions about integration decisions and workflows. API schemas and game details remain in the MCP guides and generated catalog. Add related developer workflows to this plugin as needed; creator-facing live actions or internal administration require their own product and permission design.

Platform terminology and Contest rules come from `qaragon://guides/domain-language` and `qaragon://guides/contest-rules`. The integration skill directs work to the relevant glossary sections and configuration/result work to the rules. The MCP integration prompt references the glossary and includes the Contest rules. Their canonical source is Core's `docs/`; the infra `mcp:guides` task refreshes the bundled copies. Refresh the installed plugin after a skill change and deploy the updated MCP before testing its new resources through a hosted connection.
