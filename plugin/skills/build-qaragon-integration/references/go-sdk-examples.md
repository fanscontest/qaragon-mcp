# Go SDK examples

Call `get_sdk_setup` with `language: "go"` for current installation guidance. These server-side examples use the public `github.com/fanscontest/platform-sdk-go` package. Confirm the installed release's generated method names against the operation returned by `describe_operation`.

## List Spaces for an authorized identity

Create the client once with the configured API origin (for example, the staging origin for a staging tenant). The BFF obtains the tenant key from its secret configuration and resolves `authorizedPIID` from its authenticated user mapping. Set the acting identity per request so concurrent users do not share a mutable identity header.

```go
package integration

import (
    "context"

    platform "github.com/fanscontest/platform-sdk-go"
)

func NewClient(apiOrigin string) *platform.APIClient {
    configuration := platform.NewConfiguration()
    configuration.Servers = platform.ServerConfigurations{{URL: apiOrigin}}
    return platform.NewAPIClient(configuration)
}

func ListSpaces(ctx context.Context, client *platform.APIClient, tenantKey, authorizedPIID string) (*platform.SpacesSpaceListResponse, error) {
    ctx = context.WithValue(ctx, platform.ContextAccessToken, tenantKey)
    page, _, err := client.SpacesAPI.GetSpaces(ctx).
        XActingAs(authorizedPIID).
        Limit(20).
        Execute()
    return page, err
}
```

`GET /v2/spaces` returns the envelope modeled by `SpacesSpaceListResponse`. Preserve its pagination when adapting it to the app. The key and acting identity stay on the BFF side of the request path.

## Verify a webhook

Use the SDK's signature-only helper before decoding the complete current envelope. That keeps the contract's batch/delivery fields available to the tenant's processing and deduplication code.

```go
package integration

import (
    "time"

    platform "github.com/fanscontest/platform-sdk-go"
)

func VerifyWebhook(signingSecret, signatureHeader string, rawBody []byte, now time.Time) error {
    return platform.VerifyWebhookSignature(
        signingSecret, signatureHeader, rawBody,
        platform.DefaultWebhookTolerance, now,
    )
}
```

The caller supplies the subscription's signing secret, the received `X-FC-Signature` header and the unmodified body bytes. Only decode/process the envelope after this returns nil. Use the event's current MCP schema to decide which identifiers and data to store, and acknowledge only after durable processing intent is recorded. This helper does not itself acknowledge or deduplicate deliveries.
