// Package catalog loads the public OpenAPI contract into a searchable index.
package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

var methods = []string{"get", "post", "put", "patch", "delete", "options", "head", "trace"}

// Operation is one HTTP operation in the tenant-facing contract.
type Operation struct {
	ID          string         `json:"operation_id"`
	Method      string         `json:"method"`
	Path        string         `json:"path"`
	Summary     string         `json:"summary,omitempty"`
	Description string         `json:"description,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Details     map[string]any `json:"details"`
	Schemas     map[string]any `json:"schemas,omitempty"`
	searchText  string
}

// Event is a public tenant webhook event type.
type Event struct {
	Type        string         `json:"type"`
	Description string         `json:"description"`
	DataSchema  map[string]any `json:"data_schema"`
	ExampleData map[string]any `json:"example_data"`
}

type webhookDocument struct {
	EnvelopeSchema map[string]any `json:"envelope_schema"`
	Events         []Event        `json:"events"`
}

// Catalog contains the OpenAPI document and its tenant-facing operations.
type Catalog struct {
	Document       map[string]any
	Operations     []Operation
	Events         []Event
	EnvelopeSchema map[string]any
	byID           map[string]Operation
	byEventType    map[string]Event
	GlobalSecurity any
}

// Load reads the assembled public contract and the source-derived webhook catalog.
func Load(openAPIPath, eventsPath string) (*Catalog, error) {
	contract, err := readObject(openAPIPath)
	if err != nil {
		return nil, fmt.Errorf("read OpenAPI contract: %w", err)
	}
	eventsDocument, err := os.ReadFile(eventsPath)
	if err != nil {
		return nil, fmt.Errorf("read webhook event catalog: %w", err)
	}
	var webhookEvents webhookDocument
	if err := json.Unmarshal(eventsDocument, &webhookEvents); err != nil {
		return nil, fmt.Errorf("decode webhook event catalog: %w", err)
	}
	if len(webhookEvents.Events) == 0 {
		return nil, fmt.Errorf("webhook event catalog is empty")
	}
	if len(webhookEvents.EnvelopeSchema) == 0 {
		return nil, fmt.Errorf("webhook event catalog has no envelope schema")
	}

	c := &Catalog{
		Document:       contract,
		Events:         webhookEvents.Events,
		EnvelopeSchema: webhookEvents.EnvelopeSchema,
		byID:           make(map[string]Operation),
		byEventType:    make(map[string]Event, len(webhookEvents.Events)),
		GlobalSecurity: contract["security"],
	}
	for _, event := range c.Events {
		if event.Type == "" {
			return nil, fmt.Errorf("webhook event catalog contains an empty event type")
		}
		if event.Description == "" || len(event.DataSchema) == 0 || event.ExampleData == nil {
			return nil, fmt.Errorf("webhook event %q is missing its description, data schema, or example", event.Type)
		}
		if _, exists := c.byEventType[event.Type]; exists {
			return nil, fmt.Errorf("webhook event catalog contains duplicate event type %q", event.Type)
		}
		c.byEventType[event.Type] = event
	}
	paths, ok := contract["paths"].(map[string]any)
	if !ok || len(paths) == 0 {
		return nil, fmt.Errorf("OpenAPI contract has no paths")
	}
	schemas := schemaMap(contract)
	for path, rawPathItem := range paths {
		pathItem, ok := rawPathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, method := range methods {
			rawOperation, exists := pathItem[method]
			if !exists {
				continue
			}
			details, ok := rawOperation.(map[string]any)
			if !ok {
				continue
			}
			id := stringValue(details["operationId"])
			if id == "" {
				return nil, fmt.Errorf("operation %s %s has no operationId", strings.ToUpper(method), path)
			}
			if _, exists := c.byID[id]; exists {
				return nil, fmt.Errorf("duplicate operationId %q", id)
			}
			merged := cloneMap(details)
			if pathParameters, ok := pathItem["parameters"]; ok {
				if operationParameters, ok := merged["parameters"].([]any); ok {
					merged["parameters"] = append(cloneSlice(pathParameters), operationParameters...)
				} else {
					merged["parameters"] = pathParameters
				}
			}
			operation := Operation{
				ID:          id,
				Method:      strings.ToUpper(method),
				Path:        path,
				Summary:     stringValue(details["summary"]),
				Description: stringValue(details["description"]),
				Tags:        stringSlice(details["tags"]),
				Details:     merged,
				Schemas:     referencedSchemas(merged, schemas),
			}
			operation.searchText = strings.ToLower(strings.Join([]string{
				operation.ID,
				operation.Method,
				operation.Path,
				operation.Summary,
				operation.Description,
				strings.Join(operation.Tags, " "),
			}, " "))
			c.Operations = append(c.Operations, operation)
			c.byID[id] = operation
		}
	}
	if len(c.Operations) == 0 {
		return nil, fmt.Errorf("OpenAPI contract contains no operations")
	}
	sort.Slice(c.Operations, func(i, j int) bool {
		if c.Operations[i].Path == c.Operations[j].Path {
			return c.Operations[i].Method < c.Operations[j].Method
		}
		return c.Operations[i].Path < c.Operations[j].Path
	})
	sort.Slice(c.Events, func(i, j int) bool { return c.Events[i].Type < c.Events[j].Type })
	return c, nil
}

// Search returns at most limit operations containing every query term.
func (c *Catalog) Search(query string, limit int) []Operation {
	terms := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	if limit < 1 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}
	results := make([]Operation, 0, limit)
	for _, operation := range c.Operations {
		matches := true
		for _, term := range terms {
			if !strings.Contains(operation.searchText, term) {
				matches = false
				break
			}
		}
		if matches {
			results = append(results, operation)
			if len(results) == limit {
				break
			}
		}
	}
	return results
}

// Operation returns the operation matching the exact operationId.
func (c *Catalog) Operation(id string) (Operation, bool) {
	operation, ok := c.byID[id]
	return operation, ok
}

// Event returns a public tenant webhook event by its exact event type.
func (c *Catalog) Event(eventType string) (Event, bool) {
	event, ok := c.byEventType[eventType]
	return event, ok
}

// OperationCount returns the number of indexed HTTP operations.
func (c *Catalog) OperationCount() int { return len(c.Operations) }

// DocumentJSON returns the assembled public contract as formatted JSON.
func (c *Catalog) DocumentJSON() ([]byte, error) {
	return json.MarshalIndent(c.Document, "", "  ")
}

func readObject(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func schemaMap(document map[string]any) map[string]any {
	components, _ := document["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	return schemas
}

func referencedSchemas(value any, schemas map[string]any) map[string]any {
	result := make(map[string]any)
	var walk func(any, int)
	walk = func(node any, depth int) {
		if depth > 24 {
			return
		}
		switch typed := node.(type) {
		case map[string]any:
			if ref := stringValue(typed["$ref"]); strings.HasPrefix(ref, "#/components/schemas/") {
				name := strings.TrimPrefix(ref, "#/components/schemas/")
				if _, found := result[name]; !found {
					if schema, exists := schemas[name]; exists {
						result[name] = schema
						walk(schema, depth+1)
					}
				}
			}
			for _, child := range typed {
				walk(child, depth+1)
			}
		case []any:
			for _, child := range typed {
				walk(child, depth+1)
			}
		}
	}
	walk(value, 0)
	if len(result) == 0 {
		return nil
	}
	return result
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func stringSlice(value any) []string {
	values, _ := value.([]any)
	result := make([]string, 0, len(values))
	for _, item := range values {
		if text, ok := item.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func cloneMap(source map[string]any) map[string]any {
	copy := make(map[string]any, len(source))
	for key, value := range source {
		copy[key] = value
	}
	return copy
}

func cloneSlice(value any) []any {
	slice, _ := value.([]any)
	return append([]any(nil), slice...)
}
