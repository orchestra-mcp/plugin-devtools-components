package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ComponentSyncFigmaSchema returns the JSON Schema for the component_sync_figma tool.
func ComponentSyncFigmaSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Directory containing component files to sync",
			},
			"file_key": map[string]any{
				"type":        "string",
				"description": "Optional Figma file key to sync components from",
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// ComponentSyncFigma returns a handler that syncs components with a Figma file.
func ComponentSyncFigma() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		fileKey := helpers.GetString(req.Arguments, "file_key")
		if fileKey == "" {
			return helpers.TextResult("Set FIGMA_ACCESS_TOKEN and provide file_key to sync components"), nil
		}
		return helpers.TextResult("Figma component sync not yet implemented. Use integration.figma tools to get components."), nil
	}
}
