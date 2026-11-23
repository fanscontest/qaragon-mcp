package mcpserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/qaragon/mcp/internal/catalog"
)

func TestContestConceptsArePublishedAndReferencedInIntegrationPrompts(t *testing.T) {
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
	client := mcp.NewClient(&mcp.Implementation{Name: "contest-concepts-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	prompt, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      "build_tenant_integration",
		Arguments: map[string]string{"language": "Go", "goal": "Configure a Quiz contest with groups and zone deductions"},
	})
	if err != nil {
		t.Fatal(err)
	}
	text, ok := prompt.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatal("integration prompt has no text guidance")
	}
	if !strings.Contains(text.Text, "Read the relevant sections of qaragon://guides/domain-language") {
		t.Fatal("integration prompt must direct the agent to the relevant domain vocabulary")
	}
	if !strings.Contains(text.Text, "naming rules") {
		t.Fatal("integration prompt must require the canonical naming rules")
	}

	for _, guide := range []string{"domain-language", "contest-rules"} {
		t.Run(guide, func(t *testing.T) {
			resource, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: guideURIBase + guide})
			if err != nil {
				t.Fatal(err)
			}
			if len(resource.Contents) != 1 || resource.Contents[0].Text == "" {
				t.Fatal("concept guide is empty")
			}
			if guide == "domain-language" && !strings.Contains(resource.Contents[0].Text, "## Naming rules") {
				t.Fatal("published glossary must include the naming rules")
			}
			if !strings.Contains(text.Text, guideURIBase+guide) {
				t.Fatal("integration prompt omitted the published concept guide reference")
			}
		})
	}
}
