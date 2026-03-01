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

// ComponentListSchema returns the JSON Schema for the component_list tool.
func ComponentListSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Directory to search for component files",
			},
			"filter": map[string]any{
				"type":        "string",
				"description": "Optional filter string; only include files whose path contains this string",
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// ComponentList returns a handler that lists component files in a directory.
func ComponentList() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		directory := helpers.GetString(req.Arguments, "directory")
		filter := helpers.GetString(req.Arguments, "filter")

		componentExts := map[string]bool{
			".tsx":    true,
			".vue":    true,
			".svelte": true,
			".jsx":    true,
		}

		type entry struct {
			path  string
			lines int
		}
		var components []entry

		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if !componentExts[ext] {
				return nil
			}
			if filter != "" && !strings.Contains(path, filter) {
				return nil
			}
			lineCount, _ := countLines(path)
			components = append(components, entry{path: path, lines: lineCount})
			return nil
		})
		if err != nil {
			return helpers.ErrorResult("walk_error", fmt.Sprintf("failed to walk directory: %v", err)), nil
		}

		if len(components) == 0 {
			return helpers.TextResult("No component files found in " + directory), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Components in %s (%d)\n\n", directory, len(components))
		for _, c := range components {
			fmt.Fprintf(&b, "- %s (%d lines)\n", c.path, c.lines)
		}
		return helpers.TextResult(b.String()), nil
	}
}

// countLines counts the number of lines in a file.
func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}
