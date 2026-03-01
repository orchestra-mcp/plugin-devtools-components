package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ComponentCreateSchema returns the JSON Schema for the component_create tool.
func ComponentCreateSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Directory where the component file will be created",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Component name (PascalCase)",
			},
			"framework": map[string]any{
				"type":        "string",
				"description": "Target framework: react, vue, or svelte (default: react)",
				"enum":        []any{"react", "vue", "svelte"},
			},
		},
		"required": []any{"directory", "name"},
	})
	return s
}

// ComponentCreate returns a handler that creates a new component file.
func ComponentCreate() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory", "name"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		directory := helpers.GetString(req.Arguments, "directory")
		name := helpers.GetString(req.Arguments, "name")
		framework := helpers.GetStringOr(req.Arguments, "framework", "react")

		var filePath, content string
		switch framework {
		case "vue":
			filePath = filepath.Join(directory, name+".vue")
			content = vueTemplate(name)
		case "svelte":
			filePath = filepath.Join(directory, name+".svelte")
			content = svelteTemplate(name)
		default:
			filePath = filepath.Join(directory, name+".tsx")
			content = reactTemplate(name)
		}

		if err := os.MkdirAll(directory, 0755); err != nil {
			return helpers.ErrorResult("create_error", fmt.Sprintf("failed to create directory: %v", err)), nil
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return helpers.ErrorResult("create_error", fmt.Sprintf("failed to write component file: %v", err)), nil
		}

		return helpers.TextResult(fmt.Sprintf("Component created at %s", filePath)), nil
	}
}

func reactTemplate(name string) string {
	return fmt.Sprintf(`import React from 'react';

interface %sProps {
  // Add props here
}

const %s: React.FC<%sProps> = (props) => {
  return (
    <div>
      <h1>%s</h1>
    </div>
  );
};

export default %s;
`, name, name, name, name, name)
}

func vueTemplate(name string) string {
	return fmt.Sprintf(`<template>
  <div>
    <h1>%s</h1>
  </div>
</template>

<script>
export default {
  name: '%s',
  props: {
    // Add props here
  },
  data() {
    return {};
  },
};
</script>

<style scoped>
</style>
`, name, name)
}

func svelteTemplate(name string) string {
	return fmt.Sprintf(`<script>
  // Add props here
  export let title = '%s';
</script>

<div>
  <h1>{title}</h1>
</div>

<style>
</style>
`, name)
}
