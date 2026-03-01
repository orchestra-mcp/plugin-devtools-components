package internal

import (
	"github.com/orchestra-mcp/plugin-devtools-components/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// ComponentsPlugin registers all component management tools with the plugin builder.
type ComponentsPlugin struct{}

// RegisterTools registers all 6 component tools on the given plugin builder.
func (cp *ComponentsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	builder.RegisterTool("component_list",
		"List component files (.tsx, .vue, .svelte, .jsx) in a directory",
		tools.ComponentListSchema(), tools.ComponentList())

	builder.RegisterTool("component_inspect",
		"Inspect a component file: extract props, exports, and imports",
		tools.ComponentInspectSchema(), tools.ComponentInspect())

	builder.RegisterTool("component_create",
		"Create a new component file for React, Vue, or Svelte",
		tools.ComponentCreateSchema(), tools.ComponentCreate())

	builder.RegisterTool("component_preview",
		"Show the full source of a component file with syntax highlighting",
		tools.ComponentPreviewSchema(), tools.ComponentPreview())

	builder.RegisterTool("component_library",
		"Generate a markdown index of all components in a directory",
		tools.ComponentLibrarySchema(), tools.ComponentLibrary())

	builder.RegisterTool("component_sync_figma",
		"Sync components with a Figma file (requires integration.figma tools)",
		tools.ComponentSyncFigmaSchema(), tools.ComponentSyncFigma())
}
