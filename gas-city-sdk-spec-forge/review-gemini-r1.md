YOLO mode is enabled. All tool calls will be automatically approved.
Loaded cached credentials.
YOLO mode is enabled. All tool calls will be automatically approved.
I will read `PLAN.md`, `ideas-ranked.md`, and `research-notes.md` to understand the project's current state and intended direction.
Here is a detailed review of `PLAN.md` with proposed architectural improvements, new features, and refinements based on the project's goals (`ideas-ranked.md` and `research-notes.md`).

The current plan is exceptionally strong, particularly in its definition of the **Agent Runtime** abstraction and the **Progressive Capability Model**. However, it can be improved by formalizing observability (logging/streaming), enhancing the robustness of the adapter ecosystem (contract testing CLI), and broadening the isolation model.

---

### **Review Summary**

*   **Architecture:** Strong foundation. The "Runtime Adapter" pattern effectively decouples orchestration from execution.
*   **Completeness:** Covers the core lifecycle well. Missing some specifics on **cross-runtime communication** (Task Results) and **unified logging** for non-interactive agents.
*   **Robustness:** The "Contract Test" concept is great but should be exposed to users to guarantee custom adapter quality.
*   **Usability:** The `gc init` / `gc level` flow is a standout feature for adoption.

---

### **Proposed Revisions**

#### **1. Formalize `gc test-adapter` for Ecosystem Robustness**
**Rationale:** You mention contract tests in Section 15.2, but currently they look like internal Go tests. To build a robust ecosystem of adapters (e.g., community-contributed ones), users need a way to validate *their* custom adapters against your spec without writing Go test code.
**Improvement:** Expose the contract test harness as a CLI command.

**Change to Section 11.1 (Command Structure):**
```diff
  gc migrate [--dry-run]             # Generate gas-city.toml from Gas Town workspace
  gc doctor                          # Health checks (extended from gt doctor)

- gc test-adapter <name>             # Run contract tests against an adapter
+ gc test-adapter <name> [--config=c.toml]  # Run conformance suite against an adapter
+                                            # Returns 0 on success, prints failure report
```

**Change to Section 15.2 (Contract Tests):**
```diff
- Every RuntimeAdapter passes the same contract test suite:
+ The `gc test-adapter` command runs this suite against any registered adapter.
+ This allows custom adapter authors to certify compliance with the Gas City spec.
```

#### **2. Expand Isolation Modes (`none`, `directory`, `container`)**
**Rationale:** The plan currently defaults to `isolation = "worktree"`. This is Gas Town's legacy behavior. However, for a general-purpose SDK, supporting `container` (Docker) is crucial for security, and `none` is crucial for simple "helper" agents. `directory` offers a middle ground (copy of files, no git worktree overhead).
**Improvement:** Explicitly define these modes in the schema.

**Change to Section 3.2 (Full Schema):**
```diff
  ephemeral = false               # Whether this agent is created/destroyed per-task
- isolation = "worktree"          # "none", "worktree", "directory", "container"
+ isolation = "worktree"          # Isolation strategy:
+                                 # "none"      - Runs in workspace root (dangerous, simple)
+                                 # "worktree"  - Git worktree (Gas Town default)
+                                 # "directory" - Copy of files to temp dir
+                                 # "container" - Docker/OCI container (requires docker adapter)
```

#### **3. Standardize Task Result Schema (Solve Open Question #3)**
**Rationale:** Open Question #3 asks about "Multi-runtime task serialization." If a Claude coordinator sends work to a Codex worker, how does it know the result? A standardized JSON schema for *results* is as important as the schema for *tasks*.
**Improvement:** Define `TaskResult` struct and require adapters to standardize their output parsing to match it.

**Add to Section 8.1 (Task Backend Interface):**
```diff
+ type TaskResult struct {
+     ID           string         `json:"task_id"`
+     Status       TaskStatus     `json:"status"` // Completed, Failed
+     Output       string         `json:"output"` // Summary text
+     Artifacts    []string       `json:"artifacts"` // File paths created/modified
+     Metrics      AgentMetrics   `json:"metrics"`   // Cost, tokens, time
+     Error        string         `json:"error,omitempty"`
+ }
```

**Add to Section 2.2 (RuntimeAdapter Interface):**
```diff
  // Work assignment
  Assign(handle AgentHandle, task TaskDescriptor) error
+ // WaitForResult blocks until the task is done and returns the standardized result
+ WaitForResult(handle AgentHandle) (TaskResult, error) 
```

#### **4. Unified Structured Logging**
**Rationale:** The current plan relies heavily on `CaptureOutput` (reading bytes). For a multi-agent system, you need **attributed** logging (knowing *which* agent said *what*).
**Improvement:** Add a logging requirement to the RuntimeAdapter or a wrapper that tags output lines.

**Change to Section 6.3 (Built-in Consumers):**
```diff
  | Consumer | Subscribes To | Output |
  |----------|--------------|--------|
  | CLI activity feed | All events | `gc activity --follow` real-time display |
- | Structured logger | All events | JSON lines to `workspace/.gc/events.jsonl` |
+ | Structured logger | All events + logs | `workspace/.gc/logs.jsonl` (attributed with agent_id) |
```

#### **5. Runtime Auto-Detection**
**Rationale:** To make the "Level 0" experience frictionless, users shouldn't have to specify `runtime = "claude-code"` if it's the only thing installed.
**Improvement:** Add auto-detection logic to the config loader.

**Change to Section 3.2 (Full Schema):**
```diff
  [[agents]]
  name = "string"                 # Required. Agent identifier.
  runtime = "string"              # Runtime adapter: "claude-code", "codex", ...
-                                 # Omit for auto-detection.
+                                 # If Omitted:
+                                 # 1. Check for `claude` binary -> "claude-code"
+                                 # 2. Check for `docker` binary -> "docker"
+                                 # 3. Fallback -> "subprocess"
```

#### **6. Refined `gc migrate` Report**
**Rationale:** Migration is the scariest part for existing users. The plan mentions `gc migrate --dry-run`, but specifying the *output format* builds confidence.
**Improvement:** Describe the "Migration Report" explicitly.

**Change to Section 12.1 (Migration Algorithm):**
```diff
+ // Report generates a human-readable diff of what will change
+ func (m *Migrator) Report() string {
+    // 1. Lists mapped agents (Gas Town Role -> Gas City Agent)
+    // 2. Lists config files sourced (town.json, rigs.json, etc.)
+    // 3. Flags "manual intervention needed" items (e.g., custom hooks)
+    // 4. Shows the resulting gas-city.toml content
+ }
```

#### **7. Event Bus "Replay" Clarification**
**Rationale:** Section 6.2 mentions a buffer of 10,000 events. It's important to clarify that this is for *new* subscribers (like a dashboard connecting late) to get context, not for permanent storage.
**Improvement:** Explicitly state the replay policy.

**Change to Section 6.2 (Bus Implementation):**
```diff
  type EventBus struct {
      mu          sync.RWMutex
      subscribers []chan<- Event
-     buffer      *ring.Buffer[Event]  // Last 10,000 events for replay
+     buffer      *ring.Buffer[Event]  // Ring buffer (10k events) for "catch-up" on subscribe
  }
  
  func (b *EventBus) Subscribe() <-chan Event {
      ch := make(chan Event, 100)
+     // Replay buffered events to new subscriber immediately
+     for _, event := range b.buffer.All() {
+         ch <- event
+     }
      b.mu.Lock()
      // ...
```

### **Formal Git-Diff of Critical Changes**

Here is the combined view of the most critical structural changes to `PLAN.md`.

```diff
--- PLAN.md
+++ PLAN.md
@@ -53,6 +53,7 @@
     // Work assignment
     Assign(handle AgentHandle, task TaskDescriptor) error
+    WaitForResult(handle AgentHandle) (TaskResult, error) // Standardized result
     Nudge(handle AgentHandle, message string) error
     GetState(handle AgentHandle, error)
 
@@ -107,6 +108,16 @@
     LastPingLatency time.Duration
     MemoryUsageMB   float64 // 0 if not measurable
 }
+
+// TaskResult normalizes completion data across runtimes
+type TaskResult struct {
+    ID           string
+    Status       TaskStatus     // Completed, Failed
+    Output       string         // Text summary
+    Artifacts    []string       // Files changed
+    Metrics      AgentMetrics
+    Error        string
+}
 
 // TaskDescriptor describes a task to assign to an agent.
 type TaskDescriptor struct {
@@ -194,7 +205,7 @@
 role = "worker"                 # "coordinator", "supervisor", "observer", ...
 scope = "project"               # "workspace" (one total) or "project" (one per project)
 ephemeral = false               # Whether this agent is created/destroyed per-task
-isolation = "worktree"          # "none", "worktree", "directory", "container"
+isolation = "worktree"          # "none", "worktree", "directory", "container"
 
 [agents.runtime_config]         # Adapter-specific settings
 command = "claude"              # Command to run
@@ -582,6 +593,7 @@
 gc validate                        # Config validation with diagnostics
 gc migrate [--dry-run]             # Generate gas-city.toml from Gas Town workspace
 gc doctor                          # Health checks (extended from gt doctor)
+gc test-adapter <name>             # Run contract tests against an adapter
 
-gc test-adapter <name>             # Run contract tests against an adapter
-
 gc config show                     # Display resolved config
 gc version
```
