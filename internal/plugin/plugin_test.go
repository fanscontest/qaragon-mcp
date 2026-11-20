package plugin_test

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fanscontest/qaragon-mcp/internal/catalog"
	"github.com/fanscontest/qaragon-mcp/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMarketplaceResolvesInstallablePlugin(t *testing.T) {
	var marketplace struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name   string `json:"name"`
			Source struct {
				Path string `json:"path"`
			} `json:"source"`
		} `json:"plugins"`
	}
	readJSON(t, "../../.agents/plugins/marketplace.json", &marketplace)
	if marketplace.Name != "qaragon" || len(marketplace.Plugins) != 1 {
		t.Fatalf("expected one Qaragon plugin: %+v", marketplace)
	}
	entry := marketplace.Plugins[0]
	root := filepath.Join("../..", entry.Source.Path)
	var manifest struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		Extensions map[string]struct {
			Interface struct {
				DisplayName string `json:"displayName"`
			} `json:"interface"`
		} `json:"extensions"`
	}
	readJSON(t, filepath.Join(root, "plugin.json"), &manifest)
	if entry.Name != manifest.Name || manifest.Version == "" || manifest.Extensions["com.openai"].Interface.DisplayName != "Qaragon Developer" {
		t.Fatalf("marketplace and plugin identity disagree: %+v", manifest)
	}
	skills, err := filepath.Glob(filepath.Join(root, "skills", "*", "SKILL.md"))
	if err != nil || len(skills) != 1 {
		t.Fatalf("expected one discoverable integration skill: %v, %v", skills, err)
	}
}

func TestPluginReferencesResolveInPackageOrMCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	api := &catalog.Catalog{Document: map[string]any{"openapi": "3.0.3", "info": map[string]any{"title": "test", "version": "1"}, "paths": map[string]any{}}}
	server := mcpserver.New(api, "../../content/guides")
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "plugin-package-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	links := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	root := "../../plugin"
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range links.FindAllStringSubmatch(string(body), -1) {
			target := match[1]
			switch {
			case strings.HasPrefix(target, "qaragon://"):
				if _, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: target}); err != nil {
					t.Errorf("%s references unavailable MCP resource %s: %v", path, target, err)
				}
			case strings.HasPrefix(target, "https://"), strings.HasPrefix(target, "#"):
				continue
			default:
				file := filepath.Join(filepath.Dir(path), strings.SplitN(target, "#", 2)[0])
				if _, err := os.Stat(file); err != nil {
					t.Errorf("%s references a file missing from its installed package: %s", path, target)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginConnectsToPublicDocumentationWithoutCredentials(t *testing.T) {
	var configuration struct {
		Servers map[string]struct {
			Type    string            `json:"type"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	readJSON(t, "../../plugin/mcp.json", &configuration)
	server, ok := configuration.Servers["qaragon"]
	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(configuration.Servers) != 1 || !ok || server.Type != "streamable-http" || endpoint.Scheme != "https" || endpoint.Path != "/mcp" || endpoint.User != nil || endpoint.RawQuery != "" || len(server.Headers) != 0 {
		t.Fatalf("expected one credential-free connection to public Qaragon documentation: %+v", configuration)
	}
}

func readJSON(t *testing.T, path string, target any) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, target); err != nil {
		t.Fatal(err)
	}
}
