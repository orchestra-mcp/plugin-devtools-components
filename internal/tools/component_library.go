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

// ComponentLibrarySchema returns the JSON Schema for the component_library tool.
func ComponentLibrarySchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Directory to scan for component files",
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// ComponentLibrary returns a handler that generates a markdown index of all components.
func ComponentLibrary() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		directory := helpers.GetString(req.Arguments, "directory")

		componentExts := map[string]bool{
			".tsx":    true,
			".vue":    true,
			".svelte": true,
			".jsx":    true,
		}

		type entry struct {
			name  string
			file  string
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
			base := filepath.Base(path)
			stem := strings.TrimSuffix(base, ext)
			lineCount, _ := countLines(path)
			components = append(components, entry{name: stem, file: path, lines: lineCount})
			return nil
		})
		if err != nil {
			return helpers.ErrorResult("walk_error", fmt.Sprintf("failed to walk directory: %v", err)), nil
		}

		if len(components) == 0 {
			return helpers.TextResult(fmt.Sprintf("# Component Library\n\nNo components found in %s\n", directory)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "# Component Library\n\n")
		fmt.Fprintf(&b, "Directory: `%s`\n\n", directory)
		fmt.Fprintf(&b, "| Component | File | Lines |\n")
		fmt.Fprintf(&b, "|-----------|------|-------|\n")
		for _, c := range components {
			fmt.Fprintf(&b, "| %s | %s | %d |\n", c.name, c.file, c.lines)
		}
		fmt.Fprintf(&b, "\n**Total:** %d components\n", len(components))

		return helpers.TextResult(b.String()), nil
	}
}
