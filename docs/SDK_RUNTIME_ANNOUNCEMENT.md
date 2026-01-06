# Gas Town SDK Integration: Programmatic Access to Multi-Agent Orchestration

*Enabling REST/WebSocket access to Gas Town's agent management capabilities*

---

## Introduction

Gas Town has always been about orchestrating teams of AI agents—polecats working on issues, witnesses monitoring health, refineries handling merge queues. Until now, this orchestration was deeply tied to tmux terminals: agents lived in tmux sessions, operators attached to terminals to observe progress, and the entire system required a terminal multiplexer to function.

Today we're announcing the **Gas Town SDK Integration**, a new abstraction layer that decouples agent management from any specific runtime. This opens the door to headless operation, programmatic access via REST/WebSocket, and entirely new deployment models for autonomous agent teams—all while respecting your existing Claude Code authentication.

---

## The Architecture

### The Runtime Abstraction Layer

At the heart of this integration is the `AgentRuntime` interface—a unified API for managing agent lifecycle regardless of the underlying implementation.

```go
type AgentRuntime interface {
    // Lifecycle
    Start(ctx context.Context, opts StartOptions) (*AgentSession, error)
    Stop(ctx context.Context, sessionID string, force bool) error
    Restart(ctx context.Context, sessionID string, opts StartOptions) (*AgentSession, error)

    // Communication
    SendPrompt(ctx context.Context, sessionID string, prompt string) error
    StreamResponses(ctx context.Context, sessionID string) (<-chan Response, error)

    // Monitoring
    IsRunning(ctx context.Context, sessionID string) (bool, error)
    GetStatus(ctx context.Context, sessionID string) (*AgentStatus, error)
    ListSessions(ctx context.Context, filter SessionFilter) ([]AgentSession, error)

    // Capabilities
    Capabilities() RuntimeCapabilities
}
```

Every operation that Gas Town performs on agents—spawning a polecat, nudging a stuck worker, capturing output for diagnostics—now flows through this interface.

### Two Runtimes, One Interface

**TmuxRuntime** preserves all existing Gas Town behavior. It wraps the tmux package, maintains session state, applies visual themes based on rig assignment, and supports the full range of terminal-based operations like attaching for interactive debugging.

**SDKRuntime** is the new addition. By default, it spawns Claude Code CLI subprocesses that use your existing OAuth or API key authentication—the same auth you use when running `claude` directly. No separate API key configuration required.

The key insight: the same high-level operations (`Start`, `SendPrompt`, `Stop`) work identically regardless of which runtime you're using. A polecat spawned via SDK behaves the same as one in tmux—it just doesn't have a visual terminal to attach to.

---

## User Journeys

### Journey 1: The Platform Engineer

**Scenario**: Maya is building an internal platform that lets her team dispatch agents to handle support tickets. She wants API access, not terminal windows.

**The Old Way**: Maya would need to set up a machine with tmux, have her platform SSH in, run `gt` commands, and somehow scrape terminal output to understand what agents were doing.

**The New Way**:

```bash
# Start the API server with SDK runtime (uses existing Claude Code auth)
gt serve --runtime sdk --addr :8080
```

That's it. No API key configuration needed—the SDK runtime spawns `claude` CLI subprocesses that use Maya's existing OAuth authentication.

Maya's platform can now:

1. **Create sessions** via `POST /sessions`:
   ```json
   {
     "agent_id": "support-bot/handler-1",
     "role": "polecat",
     "rig_name": "support",
     "worker_name": "handler-1",
     "system_prompt": "You are a support ticket handler..."
   }
   ```

2. **Connect WebSocket** at `/sessions/{id}/ws` to receive streaming responses:
   ```javascript
   const ws = new WebSocket('ws://localhost:8080/sessions/gt-support-handler-1/ws');
   ws.onmessage = (event) => {
     const msg = JSON.parse(event.data);
     if (msg.type === 'text') {
       updateUI(msg.content);
     }
   };
   ```

3. **Send work** via the WebSocket connection (or POST):
   ```javascript
   // Send prompt through WebSocket
   ws.send(JSON.stringify({ prompt: "Please investigate ticket #4521: User can't log in" }));

   // Or via REST (but WebSocket must be connected first to receive responses)
   // POST /sessions/{id}/prompt with {"prompt": "..."}
   ```

4. **Monitor status** via `GET /sessions/{id}`:
   ```json
   {
     "session": {"session_id": "gt-support-handler-1", "running": true},
     "health": "healthy",
     "activity": {"activity_state": "active", "idle_duration": "2s"},
     "sdk_info": {"tokens_used": 3420, "turn_count": 5}
   }
   ```

Maya never sees a terminal. Her platform has full programmatic control.

**Important**: Always connect the WebSocket *before* sending prompts. Otherwise, responses generated before the WebSocket connects are lost.

---

### Journey 2: The Ops Team Running Hybrid Deployments

**Scenario**: Carlos runs the Gas Town deployment for a large engineering org. Some teams want visual terminals for debugging; others want pure API access. He needs both.

**The Solution**: Carlos runs two API servers:

```bash
# Terminal-based runtime for teams that want to attach
gt serve --runtime tmux --addr :8080

# Headless runtime for platform integrations
gt serve --runtime sdk --addr :8081
```

Both expose the same REST API. Both use the existing Claude Code authentication. Teams choose their endpoint based on their needs. The SDK server handles automated workflows; the tmux server supports interactive debugging sessions.

When a polecat gets stuck in the SDK runtime, Carlos can check its conversation history:

```bash
curl http://localhost:8081/sessions/gt-myrig-worker1/output
```

This returns the conversation transcript rather than terminal output—but the same debugging workflow applies.

---

### Journey 3: The Agent Developer Testing Locally

**Scenario**: Priya is developing a new agent behavior. She wants fast iteration without waiting for tmux sessions to spin up.

**The Solution**: She uses the SDK runtime locally:

```typescript
// examples/sdk_hello_ts/hello.ts
import WebSocket from "ws";

const API_BASE = "http://localhost:8080";

// 1. Create a session
const session = await fetch(`${API_BASE}/sessions`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    agent_id: "test/dev",
    role: "polecat",
    system_prompt: "You are a test agent. Be concise."
  })
}).then(r => r.json());

// 2. Connect WebSocket BEFORE sending any prompts
const ws = new WebSocket(`ws://localhost:8080/sessions/${session.session_id}/ws`);

ws.on("message", (data) => {
  const msg = JSON.parse(data.toString());
  if (msg.type === "text") process.stdout.write(msg.content);
  if (msg.type === "complete") console.log("\n--- Done ---");
});

// 3. Wait for connection, then send prompt via WebSocket
ws.on("open", () => {
  ws.send(JSON.stringify({ prompt: "Write hello world in Ada" }));
});
```

Priya gets streaming responses in her terminal without any tmux overhead. The SDK runtime uses her existing Claude Code OAuth session—the same one she uses for interactive development.

---

### Journey 4: The CI/CD Pipeline

**Scenario**: DevOps wants to run agent-based code review as part of their pull request workflow.

**The Solution**: The CI job starts an SDK session, sends the PR diff as a prompt, collects the response, and posts it as a comment:

```yaml
# .github/workflows/agent-review.yml
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - name: Start Gas Town API
        run: |
          gt serve --runtime sdk --addr :8080 &
          sleep 2

      - name: Run Agent Review
        run: |
          # Create session
          SESSION=$(curl -s -X POST http://localhost:8080/sessions \
            -H "Content-Type: application/json" \
            -d '{"agent_id":"ci/reviewer","role":"crew"}' | jq -r '.session_id')

          # Send diff for review
          curl -X POST http://localhost:8080/sessions/$SESSION/prompt \
            -H "Content-Type: application/json" \
            -d "{\"prompt\": \"Review this diff:\n$(git diff origin/main)\"}"

          # Capture output
          sleep 30
          curl http://localhost:8080/sessions/$SESSION/output | jq -r '.output'
```

No display server, no tmux—just pure API calls in a CI container. The Claude Code CLI handles authentication via whatever method is configured in the CI environment.

---

## Edge Cases and Error Handling

### Edge Case 1: Session Already Exists

**What happens**: You try to create a session with an ID that's already active.

```bash
curl -X POST http://localhost:8080/sessions \
  -d '{"agent_id":"test","role":"polecat","rig_name":"myrig","worker_name":"toast"}'

# Response (500 Internal Server Error):
{"error": "session already exists: gt-myrig-toast"}
```

**Why it matters**: Session IDs follow Gas Town conventions (`gt-{rig}-{worker}`). The runtime enforces uniqueness because having two agents with the same identity would cause routing chaos.

**What to do**: Either stop the existing session first, or use a different worker name.

---

### Edge Case 2: Max Concurrent Sessions Reached

**What happens**: The SDK runtime has a configurable concurrency limit (default: 10). When exceeded:

```bash
# Response (500 Internal Server Error):
{"error": "max concurrent sessions reached (10)"}
```

**Why it matters**: Each SDK session spawns a Claude Code subprocess and maintains conversation state. The semaphore prevents resource exhaustion.

**What to do**: Either increase `MaxConcurrentSessions` in the SDK config, wait for existing sessions to complete, or stop idle sessions.

---

### Edge Case 3: Session Not Found

**What happens**: You reference a session that doesn't exist (never created, or already stopped):

```bash
curl http://localhost:8080/sessions/gt-noexist-fake

# Response:
{
  "session": {"session_id": "gt-noexist-fake", "running": false},
  "health": "unknown"
}
```

**Design decision**: Rather than returning 404, we return a status object with `running: false` and `health: unknown`. This simplifies client logic—you can always call GetStatus and check the response rather than handling HTTP errors differently.

---

### Edge Case 4: WebSocket Timing

The correct order is critical:

1. **Create session** via `POST /sessions`
2. **Connect WebSocket** at `/sessions/{id}/ws`
3. **Send prompts** via WebSocket or REST

**Connecting too early**: If you connect to the WebSocket before the session exists, the connection upgrades successfully but no messages arrive. Prompts sent via the WebSocket fail silently.

**Sending prompts before WebSocket connects**: If you send a prompt via `POST /sessions/{id}/prompt` before the WebSocket is connected, responses are generated but have nowhere to go—they're dropped.

**Best practice**: Use the WebSocket for sending prompts too. Wait for the `open` event before sending:

```javascript
const session = await createSession();
const ws = new WebSocket(`ws://localhost:8080/sessions/${session.session_id}/ws`);

ws.on("open", () => {
  ws.send(JSON.stringify({ prompt: "Start working on the task" }));
});
```

---

### Edge Case 5: Prompt Sent to Stopped Session

**What happens**: The session existed but has been stopped. You try to send a prompt:

```bash
curl -X POST http://localhost:8080/sessions/gt-myrig-toast/prompt \
  -d '{"prompt": "Are you there?"}'

# Response (500 Internal Server Error):
{"error": "session not found: gt-myrig-toast"}
```

**What to do**: Check session status before sending prompts, or handle this error by recreating the session.

---

### Edge Case 6: Graceful vs Force Stop

**What happens**: `DELETE /sessions/{id}` supports a `?force=true` parameter.

- **Without force**: The runtime attempts graceful shutdown. For tmux, it sends Ctrl-C and waits briefly. For SDK, it closes stdin and lets the Claude process exit cleanly.

- **With force**: Immediate termination. Tmux sessions are killed instantly; SDK sessions have their process killed without waiting.

**When to use force**: When a session is truly stuck and graceful shutdown hangs.

---

### Edge Case 7: Claude Code Not Installed

**What happens**: You start with `--runtime sdk` but `claude` CLI isn't in PATH:

The server starts, but session creation fails:

```bash
curl -X POST http://localhost:8080/sessions -d '{"agent_id":"test","role":"polecat"}'

# Response (500 Internal Server Error):
{"error": "failed to start claude: exec: \"claude\": executable file not found in $PATH"}
```

**What to do**: Install Claude Code CLI, or ensure it's in your PATH.

---

### Edge Case 8: Tmux Not Available

**What happens**: You start with `--runtime tmux` but tmux isn't installed:

The server starts, but session creation fails:

```bash
curl -X POST http://localhost:8080/sessions -d '{"agent_id":"test","role":"polecat"}'

# Response (500 Internal Server Error):
{"error": "creating tmux session: exec: \"tmux\": executable file not found in $PATH"}
```

**What to do**: Install tmux, or use `--runtime sdk` for headless operation.

---

### Edge Case 9: Authentication

**How auth works**: The SDK runtime spawns `claude` CLI subprocesses. These subprocesses use whatever authentication you've configured for Claude Code:

- **OAuth (Claude Max)**: If you've authenticated via `claude login`, the subprocess uses your OAuth session.
- **API Key**: If you've set `ANTHROPIC_API_KEY` in Claude Code's config, the subprocess uses that.

The Gas Town API server itself does **not** read `ANTHROPIC_API_KEY` from the environment. This is intentional—it prevents accidentally overriding your OAuth authentication when you happen to have API keys in your environment (common for Anthropic employees and developers testing multiple auth methods).

---

## Feature Deep Dive: WebSocket Streaming

The WebSocket endpoint at `/sessions/{id}/ws` is the heart of real-time interaction.

### Message Types

1. **text**: Content from the agent
   ```json
   {"type": "text", "content": "Here's the code you requested...", "timestamp": "2024-01-15T10:30:00Z"}
   ```

2. **tool_call**: Agent is invoking a tool
   ```json
   {"type": "tool_call", "content": "", "timestamp": "..."}
   ```

3. **tool_result**: Tool execution completed
   ```json
   {"type": "tool_result", "content": "", "timestamp": "..."}
   ```

4. **error**: Something went wrong
   ```json
   {"type": "error", "error": "API rate limit exceeded", "timestamp": "..."}
   ```

5. **complete**: Response finished
   ```json
   {"type": "complete", "content": "complete", "timestamp": "..."}
   ```

### Bidirectional Communication

WebSocket connections support sending prompts directly:

```javascript
ws.send(JSON.stringify({ prompt: "Now explain what you did" }));
```

This is equivalent to calling `POST /sessions/{id}/prompt` but over the existing WebSocket connection—useful for conversational flows.

### Multiple Clients

Multiple WebSocket clients can connect to the same session. All clients receive all messages (broadcast). This enables:

- A dashboard showing agent activity
- A logging service capturing responses
- A human operator observing in real-time

---

## Feature Deep Dive: Runtime Capabilities

Each runtime advertises its capabilities:

```go
// TmuxRuntime capabilities
{
    SupportsStreaming:    false,  // Polling-based (500ms intervals)
    SupportsToolCalls:    false,  // Claude Code handles tools internally
    SupportsSystemPrompt: false,  // Uses CLAUDE.md files
    SupportsAttach:       true,   // Can attach terminal
    SupportsCapture:      true,   // Can capture pane output
    SupportsConcurrency:  0,      // Unlimited (tmux manages)
}

// SDKRuntime capabilities
{
    SupportsStreaming:    true,   // Real streaming via channels
    SupportsToolCalls:    true,   // Tool calls visible in stream
    SupportsSystemPrompt: true,   // Direct system prompt
    SupportsAttach:       false,  // No terminal
    SupportsCapture:      true,   // Conversation history
    SupportsConcurrency:  10,     // Configurable limit
}
```

Clients can query capabilities to adapt their behavior—for example, only showing an "Attach" button in a UI if `SupportsAttach` is true.

---

## Feature Deep Dive: Session Identity

Gas Town has strong opinions about session naming:

| Role | Session ID Pattern | Example |
|------|-------------------|---------|
| Mayor | `hq-mayor` | `hq-mayor` |
| Deacon | `hq-deacon` | `hq-deacon` |
| Witness | `gt-{rig}-witness` | `gt-myrig-witness` |
| Refinery | `gt-{rig}-refinery` | `gt-myrig-refinery` |
| Polecat | `gt-{rig}-{worker}` | `gt-myrig-toast` |
| Crew | `gt-{rig}-crew-{worker}` | `gt-myrig-crew-alice` |

This naming convention:
- Ensures uniqueness across the deployment
- Enables filtering by rig or role
- Integrates with existing Gas Town tooling (mail routing, status displays)

When you create a session, you provide `rig_name`, `worker_name`, and `role`—the runtime generates the session ID automatically.

---

## Feature Deep Dive: Health and Activity Monitoring

Every session tracks:

- **Health**: `healthy`, `degraded`, `unhealthy`, `unknown`
- **Activity State**: `active`, `stale` (>1 min idle), `stuck` (>5 min idle)
- **Timing**: Last prompt, last response, idle duration

For SDK sessions, we also track:
- **Token usage**: Cumulative input + output tokens
- **Turn count**: Number of conversation turns

This data powers the Gas Town Deacon's health checks—agents that go too long without activity get nudged or restarted.

---

## Conclusion

The Gas Town SDK Integration transforms what was a terminal-bound orchestration system into a flexible, API-driven platform. Whether you're building internal tooling, integrating with CI/CD, or simply want faster local development, the same robust agent management is now available via REST and WebSocket.

The SDK runtime respects your existing Claude Code authentication—OAuth, API key, whatever you've configured. No separate credentials to manage.

Start experimenting:

```bash
# Headless API server (uses your existing Claude Code auth)
gt serve --runtime sdk

# Or with terminal support
gt serve --runtime tmux
```

Then hit `http://localhost:8080/health` and start building.

---

## API Reference Summary

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/sessions` | POST | Create new session |
| `/sessions` | GET | List all sessions |
| `/sessions/{id}` | GET | Get session status |
| `/sessions/{id}` | DELETE | Stop session |
| `/sessions/{id}/prompt` | POST | Send prompt |
| `/sessions/{id}/output` | GET | Capture output |
| `/sessions/{id}/ws` | GET | WebSocket streaming |
