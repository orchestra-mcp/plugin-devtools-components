package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ComponentPreviewSchema returns the JSON Schema for the component_preview tool.
func ComponentPreviewSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Directory to search for the component",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Component name to preview",
			},
		},
		"required": []any{"directory", "name"},
	})
	return s
}

// ComponentPreview returns a handler that shows the full source of a component file.
func ComponentPreview() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory", "name"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		directory := helpers.GetString(req.Arguments, "directory")
		name := helpers.GetString(req.Arguments, "name")

		componentFile, err := findComponentFile(directory, name)
		if err != nil || componentFile == "" {
			return helpers.ErrorResult("not_found", fmt.Sprintf("component %q not found in %s", name, directory)), nil
		}

		content, err := os.ReadFile(componentFile)
		if err != nil {
			return helpers.ErrorResult("read_error", fmt.Sprintf("failed to read component file: %v", err)), nil
		}

		ext := filepath.Ext(componentFile)
		lang := extensionToLanguage(ext)

		var b strings.Builder
		fmt.Fprintf(&b, "## Component: %s\n\n", name)
		fmt.Fprintf(&b, "**File:** %s\n", componentFile)
		fmt.Fprintf(&b, "**Note:** Syntax highlighting language: %s\n\n", lang)
		fmt.Fprintf(&b, "```%s\n%s\n```\n", lang, string(content))

		return helpers.TextResult(b.String()), nil
	}
}

// extensionToLanguage maps file extension to syntax highlighting language name.
func extensionToLanguage(ext string) string {
	switch ext {
	case ".tsx":
		return "tsx"
	case ".jsx":
		return "jsx"
	case ".vue":
		return "vue"
	case ".svelte":
		return "svelte"
	default:
		return "text"
	}
}
