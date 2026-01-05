// Example client demonstrating SDK runtime usage.
// Run with: ANTHROPIC_API_KEY=your-key go run examples/sdk_hello/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/runtime"
)

func main() {
	// Create SDK runtime
	rt, err := runtime.NewSDKRuntime(&config.SDKRuntimeConfig{
		APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create runtime: %v\n", err)
		os.Exit(1)
	}
	defer rt.Close()

	ctx := context.Background()

	// Start a session
	session, err := rt.Start(ctx, runtime.StartOptions{
		AgentID:      "example/hello",
		Role:         runtime.RolePolecat,
		RigName:      "example",
		WorkerName:   "hello",
		SystemPrompt: "You are a helpful programming assistant. Respond concisely.",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start session: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Started session: %s\n\n", session.SessionID)

	// Start streaming responses
	respCh, err := rt.StreamResponses(ctx, session.SessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stream: %v\n", err)
		os.Exit(1)
	}

	// Send prompt
	prompt := "Write a Hello World program in Ada. Just the code, no explanation."
	fmt.Printf("Prompt: %s\n\n", prompt)
	fmt.Println("Response:")
	fmt.Println("─────────")

	if err := rt.SendPrompt(ctx, session.SessionID, prompt); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send prompt: %v\n", err)
		os.Exit(1)
	}

	// Stream responses
	timeout := time.After(60 * time.Second)
	for {
		select {
		case resp, ok := <-respCh:
			if !ok {
				fmt.Println("\n─────────")
				fmt.Println("Stream closed")
				return
			}
			switch resp.Type {
			case runtime.ResponseText:
				fmt.Print(resp.Content)
			case runtime.ResponseComplete:
				fmt.Println("\n─────────")
				fmt.Println("Done!")
				return
			case runtime.ResponseError:
				fmt.Fprintf(os.Stderr, "\nError: %v\n", resp.Error)
				return
			}
		case <-timeout:
			fmt.Fprintf(os.Stderr, "\nTimeout waiting for response\n")
			return
		}
	}
}
