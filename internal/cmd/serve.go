// Package cmd provides CLI commands for the gt tool.
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/api"
	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/runtime"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Gas Town API server",
	Long: `Start a REST/WebSocket API server for programmatic access to Gas Town agent operations.

The SDK runtime spawns Claude Code CLI subprocesses that use your existing
OAuth/auth configuration. To use direct API calls instead, pass an api_key
in the session creation request.`,
	GroupID: GroupServices,
	RunE:    runServe,
}

var (
	serveAddr        string
	serveRuntimeType string
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "Address to listen on")
	serveCmd.Flags().StringVar(&serveRuntimeType, "runtime", "tmux", "Runtime type: tmux or sdk")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Initialize runtime based on type
	var rt runtime.AgentRuntime

	if serveRuntimeType == "sdk" {
		sdkRuntime, err := runtime.NewSDKRuntime(&config.SDKRuntimeConfig{
			Model:     "claude-sonnet-4-20250514",
			MaxTokens: 4096,
		})
		if err != nil {
			return err
		}
		rt = sdkRuntime
	} else {
		rt = runtime.NewTmuxRuntime()
	}

	server := api.NewServer(rt, serveAddr)
	return server.Start()
}
