# TypeScript SDK setup

The Qaragon TypeScript SDK package is `@fanscontest/platform-sdk`. It is distributed through GitHub Packages when published.

Configure the GitHub Packages registry for the `@fanscontest` scope using the installation instructions for the current package release, then install:

```sh
npm install @fanscontest/platform-sdk
```

Use the SDK only in a tenant server/BFF for requests that need the tenant API key. Keep the `fcn_*` key in server-side secret configuration; never bundle it into browser or mobile code. For fan-scoped operations, send `X-Acting-As` with an authorized PIID.

Package publication and registry access are managed separately from this public MCP. If the package is not available to your environment, generate a client from `qaragon://api/openapi.json` and follow the contract's authentication details, or contact the Qaragon platform maintainer for the current distribution route.
