// Command devtools-components is the entry point for the devtools.components
// plugin binary. It provides 6 MCP tools for managing component libraries.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra-mcp/plugin-devtools-components/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

func main() {
	builder := plugin.New("devtools.components").
		Version("0.1.0").
		Description("Component library manager for React, Vue, and Svelte projects").
		Author("Orchestra").
		Binary("devtools-components")

	tp := &internal.ComponentsPlugin{}
	tp.RegisterTools(builder)

	p := builder.BuildWithTools()
	p.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := p.Run(ctx); err != nil {
		log.Fatalf("devtools.components: %v", err)
	}
}
