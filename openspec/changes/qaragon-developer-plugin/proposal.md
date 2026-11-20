# Qaragon Developer plugin

## Goal

Package the existing documentation MCP and one integration skill so Qaragon and tenant developers can use the same maintained public guidance when building their BFFs and apps. Branda can exercise the first version before public directory publication.

## Design

Keep `plugin/` in the MCP repository with a portable manifest, one remote MCP entry, one skill and small references/examples. A repo marketplace enables local installation. The MCP owns the current API contracts and game guides; the skill owns the integration workflow. Tenant identity/auth, server-side credentials, notifications and rewards retain their existing boundaries. This change adds no API routes, live action tools or hosted service.

The preview uses the verified staging MCP endpoint. Production accepted initialization but lacked the current webhook-detail tool during the live check. Deploy/verify the current production catalog before changing the plugin connection for public release.

## Implementation and validation

- [x] Verify the target client's package format and the deployed MCP connection.
- [x] Add a focused integration skill, SDK examples and community/webhook recipes.
- [x] Validate marketplace resolution, packaged references and MCP resources with automated tests.
- [x] Install the actual package in an isolated Codex home and verify read-only MCP calls.
- [x] Compile/exercise the SDK examples against fixture requests and signed webhook bodies.
- [x] Document installation and representative positive/negative behavior checks.

No publication, deployment, live tenant mutation or commit is included. A real assisted Branda integration remains the behavioral acceptance step after package validation.
