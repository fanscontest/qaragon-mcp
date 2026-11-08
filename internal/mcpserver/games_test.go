package mcpserver

import (
	"context"
	"github.com/fanscontest/qaragon-mcp/internal/catalog"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/url"
	"os"
	"regexp"
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
	for _, test := range []struct{ guide, goal, want string }{
		{"word-chain", "Build qaragon://guides/word-chain", "Link Clash"},
		{"community-guess", "Build Word Chain community contests", "one final"},
		{"community-guess", "Build non-hosted Word Reveal", "final_board"},
		{"community-guess", "Build Word Reveal", "final_board"},
		{"community-guess", "Build word_reveal", "final_board"},
		{"moment", "Build Moment image contests", "format: \"moment\""},
		{"community-guess", "Build non-hosted Word Chain", "final_board"},
		{"word-chain", "Host Word Chain", "Link Clash"},
		{"word-chain", "Host link-clash", "Link Rush"},
		{"word-chain", "Offer link-rush", "Link Clash"},
		{"word-chain", "Host word_chain", "Link Clash"},
		{"word-chain", "Build Link Rush", "Link Clash"},
		{"fans-say", "Build qaragon://guides/fans-say", "Ten in Sixty"},
		{"fans-say", "Offer seven-clues", "Ten in Sixty"},
		{"fans-say", "Offer ten-in-sixty", "Seven Clues"},
		{"fans-say", "Offer fans_say", "Ten in Sixty"},
		{"fans-feud", "Build a Feud Rush game", "Feud Showdown"},
		{"fans-say", "Offer Seven Clues", "Ten in Sixty"},
	} {
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

func TestBundledGuideLinksUseResolvableResources(t *testing.T) {
	links := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	for name, file := range guideFiles {
		body, err := os.ReadFile("../../content/guides/" + file)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range links.FindAllStringSubmatch(string(body), -1) {
			target, err := url.Parse(match[1])
			if err != nil {
				t.Errorf("guide %s: invalid link %q", name, match[1])
				continue
			}
			if !target.IsAbs() {
				t.Errorf("guide %s: relative link %q cannot resolve from a standalone MCP resource", name, match[1])
				continue
			}
			if target.Scheme == "qaragon" && match[1] != apiResourceURI {
				if _, ok := guideFiles[strings.TrimPrefix(target.Path, "/")]; target.Host != "guides" || !ok {
					t.Errorf("guide %s: unpublished MCP resource %q", name, match[1])
				}
			}
		}
	}
}
