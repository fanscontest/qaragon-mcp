package mcpserver

import (
	"context"
	"github.com/fanscontest/qaragon-mcp/internal/catalog"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"strings"
	"testing"
	"time"
)

func TestGameGuidesArePublishedAndIncludedInIntegrationPrompts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	api := &catalog.Catalog{Document: map[string]any{"openapi": "3.0.3", "info": map[string]any{"title": "test", "version": "1"}, "paths": map[string]any{}}}
	server := New(api, "../../content/guides")
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "game-guide-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	for _, test := range []struct{ guide, goal, want string }{{"fans-feud", "Build a Feud Rush game", "Feud Showdown"}, {"fans-say", "Offer Seven Clues", "Ten in Sixty"}} {
		t.Run(test.guide, func(t *testing.T) {
			resource, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: guideURIBase + test.guide})
			if err != nil {
				t.Fatal(err)
			}
			if len(resource.Contents) != 1 || !strings.Contains(resource.Contents[0].Text, "/v2/submissions") {
				t.Fatalf("missing shared submission guidance: %+v", resource)
			}
			prompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{Name: "build_tenant_integration", Arguments: map[string]string{"language": "Go", "goal": test.goal}})
			if err != nil {
				t.Fatal(err)
			}
			text, ok := prompt.Messages[0].Content.(*mcp.TextContent)
			if !ok || !strings.Contains(text.Text, test.want) {
				t.Fatalf("integration prompt omitted Game guide: %+v", prompt)
			}
		})
	}
}
