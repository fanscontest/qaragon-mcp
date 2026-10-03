// Package mcpserver exposes the public Qaragon developer catalog over MCP.
package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fanscontest/qaragon-mcp/internal/catalog"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	apiResourceURI = "qaragon://api/openapi.json"
	guideURIBase   = "qaragon://guides/"
	maxSearchQuery = 160
)

var guideFiles = map[string]string{
	"tenant-bff":     "tenant-bff.md",
	"webhooks":       "webhooks.md",
	"sdk-go":         "sdk-go.md",
	"sdk-typescript": "sdk-typescript.md",
	"public-mcp":     "public-mcp.md",
}

// New registers tools, prompts, and resources backed only by public artifacts.
func New(api *catalog.Catalog, guidesDir string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "qaragon-platform",
		Version: "1.0.0",
	}, nil)

	openAPIJSON, err := api.DocumentJSON()
	if err != nil {
		panic(fmt.Errorf("format public API contract: %w", err))
	}
	server.AddResource(&mcp.Resource{
		Name:        "Qaragon public API contract",
		Description: "Merged tenant-facing OpenAPI contract, including auth and request schemas.",
		URI:         apiResourceURI,
		MIMEType:    "application/json",
	}, func(_ context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(openAPIJSON),
		}}}, nil
	})

	guides := make(map[string]string, len(guideFiles))
	for name, filename := range guideFiles {
		content, readErr := os.ReadFile(filepath.Join(guidesDir, filename))
		if readErr != nil {
			panic(fmt.Errorf("read developer guide %q: %w", filename, readErr))
		}
		guides[name] = string(content)
		resourceName := name
		resourceContent := string(content)
		server.AddResource(&mcp.Resource{
			Name:        resourceName,
			Description: "Qaragon public tenant-developer guidance.",
			URI:         guideURIBase + resourceName,
			MIMEType:    "text/markdown",
		}, func(_ context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
				URI:      request.Params.URI,
				MIMEType: "text/markdown",
				Text:     resourceContent,
			}}}, nil
		})
	}

	searchTool := func(_ context.Context, _ *mcp.CallToolRequest, input struct {
		Query      string `json:"query" jsonschema:"Words to match against operation ID, path, summary, description, and tags."`
		MaxResults int    `json:"max_results,omitempty" jsonschema:"Maximum operations to return (1-25, default 10)."`
	}) (*mcp.CallToolResult, any, error) {
		query := strings.TrimSpace(input.Query)
		if len(query) > maxSearchQuery {
			return nil, nil, fmt.Errorf("query must be at most %d characters", maxSearchQuery)
		}
		results := api.Search(query, input.MaxResults)
		output := make([]operationSummary, 0, len(results))
		for _, operation := range results {
			output = append(output, summarize(operation))
		}
		return nil, searchOutput{Operations: output}, nil
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_api",
		Description: "Search Qaragon's public API by feature, operation name, method, path, or tags.",
	}, searchTool)

	describeTool := func(_ context.Context, _ *mcp.CallToolRequest, input struct {
		OperationID string `json:"operation_id" jsonschema:"Exact OpenAPI operationId returned by search_api."`
	}) (*mcp.CallToolResult, any, error) {
		operation, ok := api.Operation(strings.TrimSpace(input.OperationID))
		if !ok {
			return nil, nil, fmt.Errorf("operation_id not found; use search_api to find a public operation")
		}
		security := operation.Details["security"]
		if security == nil {
			security = api.GlobalSecurity
		}
		return nil, describeOutput{
			OperationID: operation.ID,
			Method:      operation.Method,
			Path:        operation.Path,
			Summary:     operation.Summary,
			Description: operation.Description,
			Tags:        operation.Tags,
			Security:    security,
			Details:     operation.Details,
			Schemas:     operation.Schemas,
		}, nil
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "describe_operation",
		Description: "Get the exact public contract for an operation, including parameters, request/response schemas, auth, and referenced models.",
	}, describeTool)

	sdkTool := func(_ context.Context, _ *mcp.CallToolRequest, input struct {
		Language string `json:"language" jsonschema:"SDK language: go or typescript."`
	}) (*mcp.CallToolResult, any, error) {
		name := "sdk-" + strings.ToLower(strings.TrimSpace(input.Language))
		content, ok := guides[name]
		if !ok {
			return nil, nil, fmt.Errorf("language must be go or typescript")
		}
		return nil, guideOutput{Language: input.Language, Instructions: content}, nil
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_sdk_setup",
		Description: "Return current Qaragon SDK installation and usage guidance for Go or TypeScript.",
	}, sdkTool)

	eventsTool := func(_ context.Context, _ *mcp.CallToolRequest, input struct {
		Query string `json:"query,omitempty" jsonschema:"Optional text to filter event types."`
	}) (*mcp.CallToolResult, any, error) {
		query := strings.ToLower(strings.TrimSpace(input.Query))
		if len(query) > maxSearchQuery {
			return nil, nil, fmt.Errorf("query must be at most %d characters", maxSearchQuery)
		}
		events := make([]catalog.Event, 0, len(api.Events))
		for _, event := range api.Events {
			if query == "" || strings.Contains(strings.ToLower(event.Type), query) {
				events = append(events, event)
			}
		}
		return nil, eventOutput{Events: events}, nil
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_webhook_events",
		Description: "List supported public tenant webhook event types. These are public event names, not internal broker topics.",
	}, eventsTool)

	server.AddPrompt(&mcp.Prompt{
		Name:        "build_tenant_integration",
		Description: "Plan a tenant BFF integration against Qaragon's public API.",
		Arguments: []*mcp.PromptArgument{
			{Name: "language", Description: "Implementation language, such as Go or TypeScript.", Required: true},
			{Name: "goal", Description: "The app capability the tenant is building.", Required: true},
		},
	}, func(_ context.Context, request *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		language := request.Params.Arguments["language"]
		goal := request.Params.Arguments["goal"]
		if len(language) > 80 || len(goal) > 600 {
			return nil, fmt.Errorf("language must be at most 80 characters and goal at most 600 characters")
		}
		text := fmt.Sprintf("Help me design a %s tenant BFF integration for this goal: %s\n\nUse only the supplied public Qaragon contract and integration guidance. The tenant BFF authenticates its own user and enforces its own permissions on every request. It calls Qaragon server-to-server using a tenant API key stored in its own secret configuration and sends X-Acting-As with the platform identity ID when acting for a fan. Never put the tenant API key in a client app or generated source code. The MCP server is documentation-only and does not make Qaragon API calls.\n\nIntegration guide:\n%s\n\nSDK guidance:\n%s",
			language, goal, guides["tenant-bff"], sdkGuideFor(language, guides))
		return &mcp.GetPromptResult{
			Description: "Tenant-owned backend integration plan grounded in Qaragon's public API contract.",
			Messages:    []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: text}}},
		}, nil
	})

	return server
}

type operationSummary struct {
	OperationID string   `json:"operation_id"`
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Summary     string   `json:"summary,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type searchOutput struct {
	Operations []operationSummary `json:"operations"`
}

type describeOutput struct {
	OperationID string         `json:"operation_id"`
	Method      string         `json:"method"`
	Path        string         `json:"path"`
	Summary     string         `json:"summary,omitempty"`
	Description string         `json:"description,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Security    any            `json:"security,omitempty"`
	Details     map[string]any `json:"details"`
	Schemas     map[string]any `json:"schemas,omitempty"`
}

type guideOutput struct {
	Language     string `json:"language"`
	Instructions string `json:"instructions"`
}

type eventOutput struct {
	Events []catalog.Event `json:"events"`
}

func summarize(operation catalog.Operation) operationSummary {
	return operationSummary{
		OperationID: operation.ID,
		Method:      operation.Method,
		Path:        operation.Path,
		Summary:     operation.Summary,
		Tags:        operation.Tags,
	}
}

func sdkGuideFor(language string, guides map[string]string) string {
	if strings.EqualFold(language, "typescript") || strings.EqualFold(language, "ts") {
		return guides["sdk-typescript"]
	}
	return guides["sdk-go"]
}
