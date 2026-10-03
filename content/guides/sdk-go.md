# Go SDK setup

The Qaragon Go SDK module is `github.com/fanscontest/platform-sdk-go`. Its generated API client follows the merged public OpenAPI contract and includes webhook verification helpers.

```sh
go get github.com/fanscontest/platform-sdk-go@latest
```

Use the package's generated configuration to set the Qaragon API base URL and bearer API key from server-side configuration. Keep the `fcn_*` key in a secret manager or environment injection; do not commit it or ship it to a client app. For fan-scoped operations, set `X-Acting-As` to the authorized PIID.

SDK publishing and package availability can change. If the module is not available to your build, use the public OpenAPI contract to generate a client or contact the Qaragon platform maintainer for the current distribution route. The MCP itself does not distribute credentials or grant package-registry access.
