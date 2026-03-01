package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ComponentInspectSchema returns the JSON Schema for the component_inspect tool.
func ComponentInspectSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Directory to search for the component",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Component name to find and inspect",
			},
		},
		"required": []any{"directory", "name"},
	})
	return s
}

// ComponentInspect returns a handler that inspects a component file.
func ComponentInspect() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
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

		f, err := os.Open(componentFile)
		if err != nil {
			return helpers.ErrorResult("read_error", fmt.Sprintf("failed to read component file: %v", err)), nil
		}
		defer f.Close()

		var props, exports, imports []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			if strings.Contains(trimmed, "PropTypes") ||
				(strings.HasPrefix(trimmed, "interface") && strings.Contains(trimmed, "Props")) ||
				(strings.HasPrefix(trimmed, "type") && strings.Contains(trimmed, "Props")) {
				props = append(props, trimmed)
			}

			if strings.HasPrefix(trimmed, "export ") {
				exports = append(exports, trimmed)
			}

			if strings.HasPrefix(trimmed, "import ") {
				imports = append(imports, trimmed)
			}
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Component: %s\n\n", name)
		fmt.Fprintf(&b, "**File:** %s\n\n", componentFile)

		fmt.Fprintf(&b, "### Imports (%d)\n", len(imports))
		for _, imp := range imports {
			fmt.Fprintf(&b, "- %s\n", imp)
		}

		fmt.Fprintf(&b, "\n### Props / Interfaces (%d)\n", len(props))
		for _, p := range props {
			fmt.Fprintf(&b, "- %s\n", p)
		}

		fmt.Fprintf(&b, "\n### Exports (%d)\n", len(exports))
		for _, e := range exports {
			fmt.Fprintf(&b, "- %s\n", e)
		}

		return helpers.TextResult(b.String()), nil
	}
}

// findComponentFile walks directory to find a component file matching the given name.
func findComponentFile(directory, name string) (string, error) {
	componentExts := []string{".tsx", ".vue", ".svelte", ".jsx"}
	var found string
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		stem := strings.TrimSuffix(base, ext)
		for _, e := range componentExts {
			if ext == e && stem == name {
				found = path
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found, nil
}
