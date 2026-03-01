package devtoolscomponents

import (
	"github.com/orchestra-mcp/plugin-devtools-components/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all component management tools to the builder.
func Register(builder *plugin.PluginBuilder) {
	cp := &internal.ComponentsPlugin{}
	cp.RegisterTools(builder)
}
