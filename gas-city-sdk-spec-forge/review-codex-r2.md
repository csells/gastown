OpenAI Codex v0.101.0 (research preview)
--------
workdir: /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge
model: gpt-5.3-codex
provider: openai
approval: never
sandbox: read-only
reasoning effort: xhigh
reasoning summaries: auto
session id: 019c5f7a-bdb8-7a51-94a8-28e71973513c
--------
user
Carefully review this entire plan for me and come up with your best revisions in terms of better architecture, new features, changed features, etc. to make it better, more robust/reliable, more performant, more compelling/useful, etc. For each proposed change, give me your detailed analysis and rationale/justification for why it would make the project better along with the git-diff style change versus the original plan. The plan is in PLAN.md in this directory. Read it fully before starting your review. Also compare against specs/gas-city-vision.md requirements in the parent directory to ensure the plan fully satisfies the vision. IMPORTANT: This is a READ-ONLY review. Do NOT modify, edit, or write any files. Your output is ONLY your review feedback written to stdout. Do not touch PLAN.md or any other file. This is ROUND 2 - the plan has already been revised based on Round 1 feedback. Focus on what is STILL missing or could be better.
mcp startup: no servers
2026-02-15T04:06:45.474096Z ERROR codex_core::rollout::list: state db missing rollout path for thread 019c4eaf-800f-7fb2-8f23-8f5c9626568f
2026-02-15T04:06:45.499167Z ERROR codex_core::rollout::list: state db missing rollout path for thread 019c5c75-63a4-72b0-94f5-41d1380ed56c
2026-02-15T04:06:45.523641Z ERROR codex_core::rollout::list: state db missing rollout path for thread 019c4eec-a7a7-7340-94b9-703777b31645




thinking
**Preparing to review files**

The user has asked me to review PLAN.md and compare it to specs in gas-city-vision.md. I will read these files using the exec_command method since I only have read access. It seems I shouldn't apply any patches. The user wants a thorough review, taking into account that I’ll list findings by severity, including issues and suggestions, but I can't modify the files themselves. I'll need to check AGENTS.md for further instructions related to this task.
thinking
**Planning for detailed comparisons**

The user has requested a detailed analysis for each proposed change, including a git-diff style comparison with the original plan. This means I need to suggest modifications to the text and show them in a diff format. I’ll start by reading PLAN.md and gas-city-vision.md, and I might also review other specs for context, but only these two are necessary. I’ll utilize the terminal tool for reading and ensure to include line numbers for accurate references while being mindful of the token limit.
thinking
**Gathering tools for analysis**

I plan to use `sed -n` and `nl -ba` to prepare my command outputs, but I'll stick with `cat` since it might be easier. The tool outputs won’t be visible to the user, so I’ll parse the data for my analysis. My focus will be on proposing diff-style changes based on the plan text, especially for Round 2 improvements. Although the user didn’t request it, I think it’s wise to read CLAUDE.md for any additional project guidelines, so I’ll read it alongside PLAN.md and gas-city-vision.md simultaneously.
thinking
**Deciding on additional reads**

I’m set to use `nl -ba` for line numbering in my commands, which should help with clarity. I’ve thought about reading `README.md` as well, but it doesn’t seem necessary for this task. Since the plan is to focus on the specified files, I'll stick with those. Now, it's time to run the tool and proceed with executing the commands as planned! This should keep things organized and efficient for the user.
codex
Reading `PLAN.md` and the vision spec (`../specs/gas-city-vision.md`) with line numbers so I can do a requirement-by-requirement gap check before proposing Round 2 improvements.
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && ls -la' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forgeexec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && (test -f CLAUDE.md && nl -ba CLAUDE.md | sed -n '1,200p'; echo '---'; nl -ba PLAN.md | sed -n '1,260p'; echo '---'; nl -ba ../specs/gas-city-vision.md | sed -n '1,260p')" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forgeexec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && (echo 'PLAN tail'; nl -ba PLAN.md | tail -n 260; echo '---'; echo 'VISION tail'; nl -ba ../specs/gas-city-vision.md | tail -n 260)" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
total 432
drwxr-xr-x@ 10 csells  staff    320 Feb 14 22:07 .
drwxr-xr-x  30 csells  staff    960 Feb 14 00:19 ..
-rw-r--r--@  1 csells  staff  92661 Feb 14 22:07 PLAN.md
-rw-r--r--@  1 csells  staff  20869 Feb 14 07:28 ideas-ranked.md
-rw-r--r--@  1 csells  staff  33286 Feb 14 07:23 research-notes.md
-rw-r--r--@  1 csells  staff  33546 Feb 14 09:51 review-codex-r1.md
-rw-r--r--@  1 csells  staff   3942 Feb 14 22:07 review-codex-r2.md
-rw-r--r--@  1 csells  staff   9965 Feb 14 08:02 review-gemini-r1.md
-rw-r--r--@  1 csells  staff   6911 Feb 14 22:03 review-gemini-r2.md
-rw-r--r--@  1 csells  staff   3490 Feb 14 22:01 spec-forge-state.json
 succeeded in 50ms:
PLAN tail
  2195	    })
  2196	
  2197	    t.Run("Start creates unique handles (P6)", func(t *testing.T) {
  2198	        h1, _ := adapter.Start(ctx, cfg)
  2199	        h2, _ := adapter.Start(ctx, cfg)
  2200	        defer adapter.Stop(h1, true)
  2201	        defer adapter.Stop(h2, true)
  2202	        require.NotEqual(t, h1.ID, h2.ID)
  2203	    })
  2204	
  2205	    // === Error path tests ===
  2206	
  2207	    t.Run("Nudge on stopped agent returns error", func(t *testing.T) {
  2208	        handle, _ := adapter.Start(ctx, cfg)
  2209	        adapter.Stop(handle, true)
  2210	        err := adapter.Nudge(handle, "should-fail")
  2211	        require.Error(t, err)
  2212	    })
  2213	
 succeeded in 51ms:
---
     1	# Gas City SDK — Technical Specification
     2	
     3	> **Version:** 0.3.0
     4	> **Date:** 2026-02-14
     5	> **Status:** Planning (Spec-Forge Stage 3, Round 1 integrated)
     6	> **Source:** Independent spec-forge pipeline — research-notes.md + ideas-ranked.md
     7	> **Review Round 1:** Codex (gpt-5.3), Gemini, Claude — feedback integrated
     8	
     9	---
    10	
    11	## 1. Executive Summary
    12	
    13	Gas City is an **orchestration-builder SDK** that extracts Gas Town's hardcoded 7-role multi-agent orchestration into a configurable toolkit. The SDK lets users compose their own agent topologies — from a single task-loop agent up to full multi-project orchestration — using a declarative TOML configuration.
    14	
    15	**The core problem:** Gas Town is monolithic. Its 7 roles (Mayor, Deacon, Witness, Refinery, Polecat, Crew, Dog) are hardcoded across 50+ packages, 396+ CLI commands, and deep tmux coupling. You can't use a subset of its capabilities, swap in different agent runtimes, or define custom topologies without forking the codebase.
    16	
    17	**The core insight:** Orchestration is composable. Work tracking, messaging, health monitoring, task dispatch, and merge processing are independent capabilities that happen to be welded together in Gas Town. Separating them into composable units, each accessible through a uniform Agent Runtime interface, creates a toolkit that can express Gas Town (and any other topology) as a configuration.
    18	
    19	**The positioning:** Gas City is the "Level 8" — for users who have outgrown any single orchestrator and want to build their own. Gas Town becomes one possible configuration of Gas City, alongside Ralph, Agent Teams, and any custom topology the user designs.


thinking
**Reviewing proposed changes**

The user asked me to compare with the vision but also emphasized focusing on what's missing or could be improved. I think it’s important to check the plan's consistency while proposing modifications. They need detailed analysis and justification for each change, as well as a "git-diff style change" for the original plan in PLAN.md. So I’ll create a list of proposals with analysis, justification, and the change snippets, prioritizing findings by severity with file and line references.
thinking
**Aligning with user requests**

I need to ensure my findings align with the user’s request. I can produce findings and propose changes, including a diff snippet for each change with file and line references from PLAN.md. Since the plan is long, I’ll gather line numbers using commands like `rg` and `nl` to pinpoint relevant sections. To inspect, I’d check the table of contents in PLAN.md and use `rg '^## '` to list the sections quickly. This will help me make sure nothing is overlooked.
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "''^## " PLAN.md | head -n 200' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
11:## 1. Executive Summary
42:## 2. The Agent Runtime Abstraction
416:## 3. Configuration Schema
592:## 4. The Three Milestone Shapes
949:## 5. Workspace Controller and Lifecycle
1114:## 6. Event Bus
1232:## 7. Health Monitoring
1321:## 8. Task System
1454:## 9. Messaging System
1509:## 10. Workflow Templates
1736:## 11. CLI Design
1831:## 12. Migration from Gas Town
1964:## 13. Identity and Addressing
2042:## 14. Lifecycle Hooks
2097:## 15. Testing Strategy
2277:## 16. Implementation Phases
2379:## 17. Exclusions with Rationale
2394:## 18. Formal Properties
2444:## 19. Open Questions

thinking
**Reviewing plan sections**

I need to go through the plan carefully and read each section properly. I've already looked at part of section 2, but I need to check the rest of it, starting from line 260 up to line 415. It's important to ensure I have all the necessary information from that section. I want to make sure I understand everything clearly, so I’ll take my time with this part.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '260,460p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
   260	    if err != nil {
   261	        return false, err
   262	    }
   263	    return containsPrompt(output), nil
   264	}
   265	
   266	func (a *ClaudeCodeAdapter) WaitReady(ctx context.Context, handle AgentHandle) error {
   267	    ticker := time.NewTicker(2 * time.Second)
   268	    defer ticker.Stop()
   269	    for {
   270	        select {
   271	        case <-ctx.Done():
   272	            return ctx.Err()
   273	        case <-ticker.C:
   274	            ready, err := a.IsReady(handle)
   275	            if err != nil {
   276	                return err
   277	            }
   278	            if ready {
   279	                return nil

thinking
**Reading and reviewing sections**

I need to carefully review the entire plan, which includes sections 3 to 14. The file is quite large at 92k bytes, so I’ll read it in segments. I’ll reference various line numbers using `sed -n '416,592p'` to systematically scan through it. I'll start with section 3, focusing on lines 416 to 750 to cover the full configuration schema and the progressive levels. Let’s keep going from line 460 onwards!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '460,750p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
   460	args = ["--dangerously-skip-permissions"]
   461	env = { KEY = "VALUE" }         # Extra env vars for the runtime
   462	resume_flag = "--resume"        # How this runtime resumes sessions
   463	resume_style = "flag"           # "flag" or "subcommand"
   464	
   465	[agents.pool]                   # Pool sizing for ephemeral agents
   466	min = 0                         # Minimum instances
   467	max = 5                         # Maximum instances
   468	idle_timeout = "5m"             # Time before idle instances are killed
   469	
   470	[agents.loop]                   # Task loop configuration
   471	enabled = false
   472	auto_execute = false            # GUPP: auto-start when work appears
   473	clear_context = false           # Clear agent context between tasks
   474	poll_interval = "10s"           # How often to check for new tasks
   475	
   476	[agents.health]                 # Health check configuration
   477	ping_timeout = "30s"
   478	stuck_threshold = "1h"
   479	consecutive_failures = 3        # Failures before restart
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '750,1030p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
   750	    for {
   751	        pm.mu.Lock()
   752	
   753	        pending := pm.taskBackend.CountPending(assignee: pm.agentName)
   754	        running := pm.countRunning()
   755	
   756	        // Scale up: spawn workers for pending tasks up to max
   757	        needed := min(pending, pm.config.Pool.Max) - running
   758	        if needed > 0 {
   759	            for i := 0; i < needed; i++ {
   760	                workerConfig := pm.makeWorkerConfig(i)
   761	                handle, err := pm.runtime.Start(ctx, workerConfig)
   762	                if err != nil {
   763	                    log.Error("pool scale-up failed", err)
   764	                    break
   765	                }
   766	                // Wait for readiness if adapter supports it
   767	                if probe, ok := pm.runtime.(ReadinessProbe); ok {
   768	                    readyCtx, cancel := context.WithTimeout(ctx, workerConfig.Health.PingTimeout)
   769	                    if err := probe.WaitReady(readyCtx, handle); err != nil {
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1030,1335p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
  1030	
  1031	```go
  1032	// Default priorities (used when depends_on is not specified)
  1033	var defaultPriority = map[string]int{
  1034	    "supervisor":  0,  // Health monitor must be first to watch everything
  1035	    "coordinator": 1,  // Dispatcher starts after monitor is watching
  1036	    "observer":    2,  // Per-project monitors
  1037	    "integrator":  2,  // Per-project merge processors
  1038	    "agent":       2,  // Persistent agents
  1039	    "service":     3,  // Background services (on demand)
  1040	    "worker":      3,  // Workers (on demand)
  1041	}
  1042	// Custom roles default to priority 2 (same as per-project agents)
  1043	```
  1044	
  1045	**Startup algorithm with DAG resolution and rollback:**
  1046	
  1047	```go
  1048	func StartWorkspace(cfg *WorkspaceConfig) error {
  1049	    // Build dependency graph from depends_on + default priorities
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1335,1765p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  1335	    if state.Status != StatusWorking {
  1336	        return false
  1337	    }
  1338	
  1339	    elapsed := time.Since(state.WorkStartedAt)
  1340	    noProgress := time.Since(state.LastProgressUpdate) > s.config.StuckThreshold
  1341	
  1342	    // Not stalled if:
  1343	    // 1. Agent has reported progress recently, OR
  1344	    // 2. Elapsed time < threshold
  1345	    if !noProgress || elapsed < s.config.StuckThreshold {
  1346	        return false
  1347	    }
  1348	
  1349	    // Final check: is the agent actually responsive?
  1350	    result, _ := s.runtime.Ping(handle)
  1351	    return result.OK  // Alive but not progressing = stalled
  1352	}
  1353	```
  1354	
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1765,2125p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
  1765	
  1766	    if firstErr != nil {
  1767	        exec.Status = WorkflowFailed
  1768	        exec.FailureReason = firstErr.Error()
  1769	        return exec, firstErr
  1770	    }
  1771	
  1772	    exec.Status = WorkflowCompleted
  1773	    exec.CompletedAt = time.Now()
  1774	    return exec, nil
  1775	}
  1776	```
  1777	
  1778	---
  1779	
  1780	## 11. CLI Design
  1781	
  1782	### 11.1 Command Structure
  1783	
  1784	The Gas City CLI is `gc`. It extends or replaces `gt` depending on configuration mode.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '2125,2315p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
  2125	```
  2126	
  2127	### 14.2 Template Variables
  2128	
  2129	| Variable | Available On | Description |
  2130	|----------|-------------|-------------|
  2131	| `{{.Agent}}` | All agent events | Agent name |
  2132	| `{{.Project}}` | Project-scoped events | Project name |
  2133	| `{{.Task}}` | Task events | Task ID |
  2134	| `{{.TaskTitle}}` | Task events | Task title |
  2135	| `{{.Error}}` | Failure events | Error message |
  2136	| `{{.Timestamp}}` | All events | ISO-8601 timestamp |
  2137	| `{{.Status}}` | State change events | New status |
  2138	
  2139	---
  2140	
  2141	## 15. Testing Strategy
  2142	
  2143	### 15.1 Unit Tests
  2144	



thinking
**Reviewing the plan**

I've gone through the entire plan, but noticed the CLAUDE.md file is missing from the repository, which was noted in our earlier command. That's okay! Now, I need to work on the Round 2 review, focusing on comparing the plan with the vision. The vision file is brief, but I must ensure the plan meets all requirements regarding the orchestration-builder toolkit, roles and teams, and sandbox support. It seems like the plan covers these points, but the plugin system needs a bit more clarity.
thinking
**Assessing plugin architecture**

I need to review Section 2.8, which talks about the adapter registry, but I notice plugin loading isn't clearly described. The executive summary mentions sandboxes, plugins, and hooks, but it seems the plan lacks a solid plugin architecture beyond just runtime adapters and registries for backends. There are open questions about the out-of-process adapter protocol and whether we should specify a plugin mechanism for new adapters. 

Also, while it seems the plan generally meets the uniform "factory worker" abstraction, I think the user is looking for revisions to enhance architecture and features to improve reliability and performance. I’ll propose some changes based on what’s still lacking or could be better!
thinking
**Exploring existing reviews**

In Round 1, feedback was integrated, and now it's time for us to propose more improvements. I should check the existing review-codex-r2.md to see what suggestions were already made. Even though the user didn't explicitly ask for this, it might help avoid repetition. The focus should be on what's missing for Round 2, so reviewing these documents will give me a better understanding of past suggestions. By doing this, I can ensure my contributions are new and valuable, so I'll read quickly!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && (echo 'review-codex-r2'; nl -ba review-codex-r2.md | sed -n '1,260p'; echo '---'; echo 'review-gemini-r2'; nl -ba review-gemini-r2.md | sed -n '1,260p')" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
review-codex-r2
     1	OpenAI Codex v0.101.0 (research preview)
     2	--------
     3	workdir: /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge
     4	model: gpt-5.3-codex
     5	provider: openai
     6	approval: never
     7	sandbox: read-only
     8	reasoning effort: xhigh
     9	reasoning summaries: auto
    10	session id: 019c5f7a-bdb8-7a51-94a8-28e71973513c
    11	--------
    12	user
    13	Carefully review this entire plan for me and come up with your best revisions in terms of better architecture, new features, changed features, etc. to make it better, more robust/reliable, more performant, more compelling/useful, etc. For each proposed change, give me your detailed analysis and rationale/justification for why it would make the project better along with the git-diff style change versus the original plan. The plan is in PLAN.md in this directory. Read it fully before starting your review. Also compare against specs/gas-city-vision.md requirements in the parent directory to ensure the plan fully satisfies the vision. IMPORTANT: This is a READ-ONLY review. Do NOT modify, edit, or write any files. Your output is ONLY your review feedback written to stdout. Do not touch PLAN.md or any other file. This is ROUND 2 - the plan has already been revised based on Round 1 feedback. Focus on what is STILL missing or could be better.
    14	mcp startup: no servers
    15	2026-02-15T04:06:45.474096Z ERROR codex_core::rollout::list: state db missing rollout path for thread 019c4eaf-800f-7fb2-8f23-8f5c9626568f
    16	2026-02-15T04:06:45.499167Z ERROR codex_core::rollout::list: state db missing rollout path for thread 019c5c75-63a4-72b0-94f5-41d1380ed56c
    17	2026-02-15T04:06:45.523641Z ERROR codex_core::rollout::list: state db missing rollout path for thread 019c4eec-a7a7-7340-94b9-703777b31645
    18	
    19	






thinking
**Evaluating review suggestions**

The previous codex r2 file isn't an actual review; it looks more like tool logs. I'll need to confirm whether it was an accidental save. The Gemini r2 review provides suggestions like plugin adapters, secrets management, cost tracking, and context management, which could be relevant for our own review. However, I should propose new improvements and not just replicate their suggestions, aiming for more robust details and corrections as needed.
thinking
**Evaluating architectural improvements**

I need to focus on potential improvements in architecture, features, performance, and reliability. 

First, the RuntimeAdapter interface seems broad, and there are potential issues surrounding WaitForResult semantics and concurrency. There's also a lack of capability advertisements and context metrics.

Next, the Start/Stop/Restart functions have confusion around thread safety and handle identity, which could lead to bugs. 

Finally, for the SubprocessAdapter, the way output capture is managed can be problematic for long sessions; it might be useful to propose a more unified output capture approach.
thinking
**Analyzing potential issues**

The current system doesn’t guarantee ordered delivery, as it uses goroutines that can effectively reorder messages, and delivery guarantees may not be essential. Additionally, the buffer ring and replay could call the subscriber's OnEvent without proper concurrency control, leading to potential blocking.

For the hooks, the HookExecutor executes them in goroutines but returns nil quickly. While this keeps the publishing process from being blocked, it also means errors might not propagate correctly.

Finally, the Task system’s filesystem backend utilizes a lock file but fails to handle scenarios where the process crashes, which could leave the lock file intact.
thinking
**Identifying issues with locking and tasks**

If a crash leaves a lock file, subsequent claims will always fail, so implementing stale lock detection using PID and timestamp is essential. Using `flock` for advisory locks would be better than using a lock file, since the current implementation with O_EXCL isn’t robust.

For the task loop, it lists tasks filtered by the agent’s name, but inconsistencies arise since the claim semantics require the assignee to be empty, while the filter checks for the agent's name. Tasks should clarify the distribution as either push or pull systems. Additionally, the PoolManager counts pending tasks, but this may also lead to inconsistencies as tasks might not be assigned.
thinking
**Identifying key issues and proposals**

I see several points of confusion in the current design. The task backend has redundancy since the Claim already sets InProgress status, which could be streamlined. Additionally, there are unresolved dependencies without a clear resolution or update mechanism. 

The workflows use a global runtime variable in a way that isn't ideal, and the messaging delivery semantics may lead to message loss. 

For the workspace controller, there's no clear protocol for connections on different OSes, which could be improved with gRPC or JSON-RPC.

Moreover, the crash recovery doesn't define how to reconnect to runtime handles, so it might need an `Adopt` method. There are many suggested enhancements, like isolating configurations, ensuring permissions, and refining observability in logging. Addressing concurrency and idempotency in messaging, versioning, and better coordination rules would also benefit the system significantly. Overall, there are clear gaps for improvement!
thinking
**Clarifying missing elements**

I’m noticing several gaps in the current setup. First, the `gc agent attach` command for the CLI is absent, although the runtime adapter has `SupportsAttach` for attach and detach functionalities. However, the `Attach` method is interactive and might block, so it might need context support.

Additionally, output capture for tmux isn't flowing well as it only snapshots with `capture-pane`. It might be better to use an event bus to stream output lines as events.

Furthermore, integrating with Git requires consideration for isolation worktrees, branch naming, cleanup, and concurrency issues. Lastly, there's a lack of detail around "batches" in the context of multi-project convoys, even though they mention `batches` being enabled and auto-tracking.
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "batches" PLAN.md' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
569:[batches]
571:auto_track = false              # Auto-create batches for multi-task dispatches
856:[batches]









thinking
**Identifying gaps in documentation**

I see that there are issues with the config schema and the Gas Town shape config. There's no description of the batch system, which is a significant oversight. It might have to do with convoy tracking, but the plan doesn’t clarify this. I should suggest either removing it from the schema or adding a section to describe it. Also, the mention of the integrator role doesn't detail behaviors or merge queue systems, which might be an important gap too.
thinking
**Examining role logic discrepancies**

I realize that to achieve equivalent behavior to the current Gas Town, there really needs to be a standard set of prompts or workflows that replicate Gas Town roles. It's possible the plan overlooks this. The claim about no hardcoded role logic in the SDK seems contradicted by the mention of specific roles like the Coordinator, which implies some form of hardcoding. Features like pool manager and supervisor patrol cycles depend on configurations, which raises questions about their role in monitoring and limits. It feels inconsistent overall!
thinking
**Reconciling roles and capabilities**

I need to think through the roles, maybe treating them as tags for built-in controllers instead. The coordinator and supervisor roles could be viewed as capabilities defined by config sections that enable those features. For instance, when tasks are dispatched, the coordinator acts more like an agent with the logic to assign tasks rather than a strict role. This makes the statement about "no hardcoded role logic" somewhat misleading. I might suggest revising the plan to clarify that roles are mainly descriptive, with certain internal capabilities linked to specific roles.
thinking
**Proposing architecture changes**

I’m considering recommending an architecture change to separate "control loops" from "agents" to avoid hardcoded role naming. Instead, I could suggest having "controllers" that operate certain agents at runtime. For example, the supervisor shouldn't be viewed as a separate AI agent; instead, it's part of the workspace controller. The implementation is in Go, and it monitors other agents, reinforcing that it’s a controller and not an AI agent. This raises questions about how Deacon is defined as a supervisor agent and the connection to Claude in Gas Town.
thinking
**Addressing architecture and implementation inconsistencies**

In examining the plan, it seems the supervisor is treated as a Go component rather than an AI agent, which is confusing. If it's a Go component that monitors via a runtime adapter, why label it as an agent in the config? This raises questions about the true role of supervisors and coordinators. It might make sense to separate these roles as system components rather than AI agents, allowing for clearer control logic without needing an AI agent for monitoring. Additionally, other areas are missing details, such as data modeling, observability metrics, and resource management considerations for output capture. I should suggest refining the architecture to enhance clarity and functionality.
thinking
**Proposing algorithm and improvement changes**

I see several areas for improvement in the plan. For instance, the StartWorkspace algorithm needs to correctly track instances, as the current implementation can miscount handles. Additionally, the StopWorkspace relies on pseudo-defined groups that aren't well specified, potentially leading to confusion.

I should propose a detailed list of improvements, like better file locking with `flock`, enhancing runtime detection, and adding an explicit plugin system. Each proposal will include an analysis, rationale, and a git-diff style change to provide clarity. Keeping the revisions manageable, around 8-12 changes, can help us improve the plan effectively.
thinking
**Refining task claim and event bus mechanisms**

I need to address some mismatches in task claim semantics in Section 4.1, particularly around task assignment models. I’m considering adding two claim modes: a pull model where agents claim unassigned tasks, and a push model where coordinators assign tasks directly to agents. 

Additionally, I should propose changes to improve the event bus concurrency by implementing a buffered channel for subscribers. This helps avoid unbounded goroutines while managing critical events more reliably, ensuring that our system doesn’t stall. These enhancements aim to streamline task claims and improve event handling.
thinking
**Proposing improvements for plans**

I need to focus on ensuring event reliability, especially for critical events, by treating the event bus like a durable log. This involves tracking subscriber acknowledgments for improved reliability and crash recovery.  

Additionally, I want to explore implementing an `Adopt` interface to manage agent restarts. I'll also propose an external plugin mechanism for runtime adapters to avoid code compilation issues.   

It’s essential to define batch tracking for tasks and address gaps in migration algorithms and error handling to ensure robust workflows. Providing resource limits and capability models will enhance efficiency. Overall, it'll be a detailed review of many proposed changes.
thinking
**Structuring proposed changes**

I need to ensure that when referencing files, they’re clickable and formatted correctly, using inline code like `PLAN.md:1023`. It’s also important to compare against `../specs/gas-city-vision.md` to identify any gaps in vision coverage.  

I’ll structure my findings based on severity, categorizing them into P0 (must fix), P1 (should fix), and P2 (nice-to-have), while providing a rationale and proposed diffs for each change. Including open questions or assumptions will help as well, and I'll verify exact line numbers for accurate diff references.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '40,120p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
    40	---
    41	
    42	## 2. The Agent Runtime Abstraction
    43	
    44	### 2.1 Design Rationale
    45	
    46	Every interaction Gas Town performs with an agent — starting, stopping, sending work, reading output, health-checking — currently flows through `internal/tmux/tmux.go`. This creates three problems:
    47	
    48	1. **Runtime lock-in**: You can't use Codex, Gemini, or Agent SDK without rewriting the session layer
    49	2. **Environment lock-in**: You can't run without tmux (containers, CI, headless servers)
    50	3. **Observation lock-in**: Monitoring depends on tmux `capture-pane` output scraping
    51	
    52	The Agent Runtime abstraction separates **what you do with an agent** from **how a specific agent implementation works**.
    53	
    54	### 2.2 The RuntimeAdapter Interface
    55	
    56	```go
    57	// RuntimeAdapter is the interface every agent runtime must implement.
    58	// It is the foundational abstraction of the Gas City SDK.
    59	type RuntimeAdapter interface {
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '118,260p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
   118	```go
   119	func (a *SubprocessAdapter) WaitForResult(handle AgentHandle) (TaskResult, error) {
   120	    cmd := a.getCmd(handle.ID)
   121	    err := cmd.Wait()  // Blocks until process exits
   122	
   123	    output, _ := a.readAllOutput(handle.ID)
   124	    artifacts := a.detectChangedFiles(handle)
   125	
   126	    result := TaskResult{
   127	        TaskID:    handle.Metadata["current_task"],
   128	        Output:    output,
   129	        Artifacts: artifacts,
   130	    }
   131	
   132	    if err != nil {
   133	        result.Status = StatusFailed
   134	        result.Error = err.Error()
   135	    } else {
   136	        result.Status = StatusCompleted
   137	    }









thinking
**Evaluating integration and adjustments**

The plan already has token and cost fields in AgentMetrics, like TokensIn, TokensOut, and EstimatedCostUSD, which seems to mean Gemini suggestion 3 is already integrated. That's a relief! Now, I need to adjust the suggestions accordingly. 

I notice there's some inconsistency in the Exclusions section regarding cost management versus cost tracking, which needs clarification. Finally, since the config has `env` and `clear_context`, I can propose changes for secrets management and context policy. The plugin architecture remains an open question, but there’s an adapter registry.
thinking
**Examining integration and architecture**

There's an open question regarding the out-of-process adapter protocol and JSON-RPC, which I think needs to be integrated into the plan. The built-in runtime adapters list includes codex, gemini-cli, agent-sdk, and docker, but we need to pay attention to section 2.4. 

I realized the "completion_mode" config is there, so our review should also incorporate the existing tokens and cost fields. Now, I want to identify key architecture issues like crash recovery claims, task assignment mismatches, and incomplete config for batches to address them thoroughly.
thinking
**Identifying key improvements**

I see that the CLI is missing plugin management and a clear workflow plan, even though it includes commands like `gc workflow list` and `gc workflow run`. I might propose including a plan here. 

For security, there’s a need for environment variable expansion and redaction. Observability is another area to improve; we need structured logging with correlation IDs and retention plans. I’ve noted some key lines for citations, like crash recovery, task loop algorithms, and event bus publish, to focus on improvements in features and reliability for the user.
thinking
**Reviewing key lines and proposals**

I’ll be citing key lines including the hook executor from line 2090, roles externalization in lines 923-930, and the default priority map using role names in lines 1031-1043. Config validation also needs attention with coordinator and supervisor limit lines 585-586. I see that the batches config is covered in lines 524-527. 

It’s essential to compare against vision requirements since parts are satisfied, but I want to note that the plan still hardcodes some roles. The biggest gap seems to be with the plugin system, which needs stronger development. Let’s propose revisions with detailed modifications to the plan.
thinking
**Crafting proposed changes**

We shouldn't edit files directly, but I can propose changes. For each modification, I’ll include a diff snippet to keep things clear without making it one big diff. Each change will reference specific lines and include an approximate context with `@@`.

I’ll structure the review into sections like Vision Alignment and Critical Gaps (P0), laying out problems with references, proposed revisions, rationale, and the code block diff. For example, for "Change 1," I’ll propose adding an optional interface for adoption. I plan to address task assignment discrepancies in "Change 2," updating the Task struct accordingly. Each suggested change will be numbered for ease of reference, focusing on plain text, preferably in markdown.
thinking
**Proposing specific changes**

I see that we need to update several aspects for a more robust design. First, I’ll maintain `Assignee` in the TaskFilter but also add `ClaimedBy`. Additionally, I’ll revise how we count pending tasks and manage task assignments to avoid inconsistencies. 

For Change 3, I want to implement a more reliable filesystem claim using `flock` to ensure locks are handled better, along with atomic writes to avoid partial data issues.

In Change 4, I plan to resolve contradictions around role externalization. I can propose either clarifying reserved roles or shifting towards capabilities to enhance configurability. I’ll define a `[monitoring]` section and adjust related checks in the configuration schema.
thinking
**Proposing structural changes**

I want to clarify that if a `supervisor` agent is set up, it's considered a runtime agent, while health monitoring should operate as a separate system service. To improve clarity, I propose renaming this service to either `HealthMonitor` or `SupervisorService`, which won't require it to act as an agent. 

I also recognize that we need to fill out the batch tracking section or consider removing it from the design to maintain plan completeness. It seems best to add a section covering batch systems. 

For the event bus, making it more scalable and failure-tolerant is essential; I’m suggesting significant improvements to manage events and subscriber interactions more efficiently. 

Lastly, I want to clarify the output streaming model, especially for logging agent outputs, and ensure that secrets management becomes more secure with config layering updates. I'll also propose enhancements to the JSON-RPC plugin protocol for better integration into our processes.
thinking
**Proposing changes for clarity**

I suggest unifying checkpoint semantics per adapter and including functionality to persist to `.gc/checkpoints`. The `AgentRegistry` persistence should be atomic and leak-proof. There's also a need to use `agent identity strings` in filenames but ensure they're properly sanitized to avoid invalid paths. 

Next, I need to address correctness issues in the pseudocode, like mismatches in signatures and returns in the EventBus.Subscribe function. I’ll also make various changes to improve the spec quality, including adding an adoption interface to support crash recovery for agents, modifying task structure, and updating the filesystem backend to use flock for better claims management. Lastly, I want to clarify monitoring enabling and role semantics related to the configuration schema. Let’s craft the changes carefully!
thinking
**Enhancing configuration robustness**

I’m thinking that the current coordinator and supervisor limits might be too restrictive, especially for users needing multiple coordinators in various shapes. Instead of hard limits, we could replace them with warnings or better yet, validate agent name uniqueness and dependencies without focusing on roles. 

I'll propose adding a new `[controllers]` section in the configuration schema, detailing how built-in controllers can be explicitly configured. This would reduce reliance on descriptive roles while enhancing flexibility. Also, I need to consider designs for batches, enhance the event bus, update output events, and streamline secrets management. Let's craft some concise diffs!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '500,535p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
   500	scope = "project"               # "workspace" (one total) or "project" (one per project)
   501	ephemeral = false               # Whether this agent is created/destroyed per-task
   502	isolation = "worktree"          # Isolation strategy:
   503	                                # "none"      - Runs in workspace root (simple, shared state)
   504	                                # "worktree"  - Git worktree (Gas Town default, branch isolation)
   505	                                # "directory" - Copy of files to temp dir (no git overhead)
   506	                                # "container" - Docker/OCI container (strongest isolation)
   507	depends_on = []                 # Agent names that must start before this one
   508	
   509	[agents.runtime_config]         # Adapter-specific settings
   510	command = "claude"              # Command to run
   511	args = ["--dangerously-skip-permissions"]
   512	env = { KEY = "${MY_SECRET}" }  # Extra env vars (supports ${VAR} expansion)
   513	resume_flag = "--resume"        # How this runtime resumes sessions
   514	resume_style = "flag"           # "flag" or "subcommand"
   515	
   516	[agents.pool]                   # Pool sizing for ephemeral agents
   517	min = 0                         # Minimum instances
   518	max = 5                         # Maximum instances
   519	idle_timeout = "5m"             # Time before idle instances are killed

thinking
I'm looking at the code snippet that includes `env = { KEY = "VALUE" }`. I'm questioning which part is correct, so I'll check the actual file content in that specific area to verify. Running `sed -n '505,520p'` seems like a plan, but I might want to skip the line numbers when I do this to confirm the content effectively. Let's see how that goes!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && sed -n '505,520p' PLAN.md" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
                                # "directory" - Copy of files to temp dir (no git overhead)
                                # "container" - Docker/OCI container (strongest isolation)
depends_on = []                 # Agent names that must start before this one

[agents.runtime_config]         # Adapter-specific settings
command = "claude"              # Command to run
args = ["--dangerously-skip-permissions"]
env = { KEY = "${MY_SECRET}" }  # Extra env vars (supports ${VAR} expansion)
resume_flag = "--resume"        # How this runtime resumes sessions
resume_style = "flag"           # "flag" or "subcommand"

[agents.pool]                   # Pool sizing for ephemeral agents
min = 0                         # Minimum instances
max = 5                         # Maximum instances
idle_timeout = "5m"             # Time before idle instances are killed


thinking
**Checking file content**

I’m looking at the file content near line 512, where I see `env = { KEY = "${MY_SECRET}" }`. I’m wondering if there might actually be two occurrences of something similar. So, I think I should search for `env = { KEY = "VALUE" }` in the PLAN.md file to confirm. That way, I can clarify whether this pattern appears multiple times or is just unique.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n \"env = \\{ KEY = \\\"VALUE\\\" \\}\" PLAN.md || true" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:

exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "context_policy|summarize|window_size" PLAN.md || true' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:

exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "workflow plan" PLAN.md || true' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
1866:gc workflow plan <template>        # Dry-run: print execution graph without running

thinking
**Updating workflow review**

I see that the plan now includes the `gc workflow plan` at line 1866, which means they integrated the Gemini suggestion. That’s great! However, we need to be cautious since some suggestions have already been included. It’s important to focus on what’s still missing from our review. I should make sure to read the latest file, keeping in mind that line numbers may have shifted due to recent additions. Let’s update our reading of these sections by gathering the top-level headings again.
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "''^## " PLAN.md | head -n 30' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
11:## 1. Executive Summary
42:## 2. The Agent Runtime Abstraction
463:## 3. Configuration Schema
643:## 4. The Three Milestone Shapes
1013:## 5. Workspace Controller and Lifecycle
1195:## 6. Event Bus
1317:## 7. Health Monitoring
1406:## 8. Task System
1543:## 9. Messaging System
1598:## 10. Workflow Templates
1835:## 11. CLI Design
1932:## 12. Migration from Gas Town
2065:## 13. Identity and Addressing
2143:## 14. Lifecycle Hooks
2208:## 15. Testing Strategy
2388:## 16. Implementation Phases
2490:## 17. Exclusions with Rationale
2505:## 18. Formal Properties
2555:## 19. Open Questions

thinking
**Planning crash recovery**

I want to capture key lines for crash recovery in the workspace controller section, specifically around line 1013. I need to re-open it around line 1030. Additionally, I'll be using the command `nl -ba` for the lines between 1013 to 1050. This should help ensure I have the right information for recovery purposes. It’s all about being prepared and covering all my bases!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1010,1065p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  1010	
  1011	---
  1012	
  1013	## 5. Workspace Controller and Lifecycle
  1014	
  1015	### 5.0 Workspace Controller
  1016	
  1017	The workspace controller is the long-lived process that hosts all control loops: the event bus, agent registry, pool managers, supervisor patrol, hook executor, and workflow executor. It is the central coordination point that all other subsystems depend on.
  1018	
  1019	```go
  1020	type WorkspaceController struct {
  1021	    config       *WorkspaceConfig
  1022	    registry     *AgentRegistry
  1023	    eventBus     *EventBus
  1024	    taskBackend  TaskBackend
  1025	    poolManagers map[string]*PoolManager
  1026	    supervisor   *Supervisor
  1027	    hookExec     *HookExecutor
  1028	    workflows    *WorkflowExecutor
  1029	    lockFile     string  // .gc/controller.lock

thinking
**Analyzing crash recovery**

I'm focusing on the crash recovery paragraph at line 1043, where I see that the default priority map remains role-based. Those lines will be useful for change 4. Now, I need to take a closer look at the config schema related to monitoring comments, particularly around line numbers 567 and on. I plan to open the section from lines 540 to 590 for a more thorough examination.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '520,605p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
   520	
   521	[agents.loop]                   # Task loop configuration
   522	enabled = false
   523	auto_execute = false            # GUPP: auto-start when work appears
   524	clear_context = false           # Clear agent context between tasks
   525	poll_interval = "10s"           # How often to check for new tasks
   526	
   527	[agents.health]                 # Health check configuration
   528	ping_timeout = "30s"
   529	stuck_threshold = "1h"
   530	consecutive_failures = 3        # Failures before restart
   531	kill_cooldown = "5m"            # Time between forced restarts
   532	
   533	[agents.prompts]                # Prompt templates
   534	system = "prompts/agent.md"     # Path to system prompt (Go template)
   535	nudge = "Check for work."       # Default nudge message
   536	
   537	[agents.hooks]                  # Lifecycle hooks (shell commands)
   538	on_start = ""                   # Runs after agent starts
   539	on_stop = ""                    # Runs after agent stops
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '605,635p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
   605	    if level >= 3 && cfg.Messaging != nil {
   606	        level = 4 // Level 4: Inter-Agent Messaging
   607	    }
   608	
   609	    if level >= 4 && cfg.Workflows != nil {
   610	        level = 5 // Level 5: Workflow Templates
   611	    }
   612	
   613	    if level >= 5 && hasSupervisor(cfg) {
   614	        level = 6 // Level 6: Monitoring
   615	    }
   616	
   617	    if level >= 6 && len(cfg.Projects) >= 2 && hasProjectScopedAgents(cfg) {
   618	        level = 7 // Level 7: Full Orchestration (Gas Town shape)
   619	    }
   620	
   621	    return level
   622	}
   623	```
   624	
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '635,660p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
   635	| Backend exists | `tasks.backend` in backend registry | "unknown task backend: X" |
   636	| Coordinator limit | At most 1 coordinator per workspace | "multiple coordinators not supported" |
   637	| Supervisor limit | At most 1 supervisor per workspace | "multiple supervisors not supported" |
   638	| Dependency validity | `depends_on` references existing agent names | "agent X depends on unknown agent Y" |
   639	| Dependency acyclic | `depends_on` graph has no cycles | "dependency cycle: X -> Y -> X" |
   640	
   641	---
   642	
   643	## 4. The Three Milestone Shapes
   644	
   645	### 4.1 Ralph (Level 2)
   646	
   647	A single agent with a task loop that autonomously processes queued work.
   648	
   649	```toml
   650	[workspace]
   651	name = "ralph-demo"
   652	schema_version = 1
   653	
   654	[projects.main]

thinking
**Planning algorithm gathering**

I’m thinking about gathering the Task loop algorithm section, specifically focusing on the line numbers. I want to open around lines 670 to 720 to make sure I cover the relevant parts. It's important to get these details correctly so I can assist effectively! I'll keep an eye on those specific range numbers to ensure nothing slips through the cracks. Let’s proceed!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '670,740p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
   670	
   671	**Task loop algorithm:**
   672	
   673	```
   674	LOOP:
   675	  tasks = taskBackend.ListReady(filter: {assignee: self.name, sortBy: "priority", sortOrder: "desc"})
   676	  if len(tasks) == 0:
   677	    sleep(poll_interval)
   678	    goto LOOP
   679	  task = tasks[0]  // highest priority
   680	  ok = taskBackend.Claim(task.id, self.name)  // Atomic claim
   681	  if !ok:
   682	    goto LOOP  // Another agent claimed it first
   683	  runtime.Assign(self.handle, task)
   684	  WAIT_COMPLETION:
   685	    result = runtime.WaitForResult(self.handle)  // Blocks with internal timeout
   686	    if result.Status == Completed:
   687	      taskBackend.MarkCompleted(task.id, self.name, result.Output)
   688	      if clear_context:
   689	        runtime.Stop(self.handle, graceful=true)

thinking
I need to investigate why the algorithm is still listing tasks with the filter assignee set to self.name. It seems that this might not be correct. I'll check the TaskBackend interface specifically around line 1406 to verify the claim semantics. I'll look closely at the lines around 1440, but I recall seeing something relevant at line 1367 in the old numbering. So, it makes sense to read through lines 1420 to 1470 for better context.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1406,1495p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  1406	## 8. Task System
  1407	
  1408	### 8.1 Task Backend Interface
  1409	
  1410	```go
  1411	type TaskStatus int
  1412	const (
  1413	    StatusOpen       TaskStatus = iota
  1414	    StatusReady                        // All deps resolved, available for claim
  1415	    StatusInProgress                   // Claimed and being worked on
  1416	    StatusCompleted                    // Successfully finished
  1417	    StatusFailed                       // Terminal failure
  1418	    StatusBlocked                      // Blocked on unresolved dependencies
  1419	)
  1420	
  1421	type TaskBackend interface {
  1422	    // CRUD
  1423	    List(filter TaskFilter) ([]Task, error)
  1424	    Get(id string) (Task, error)
  1425	    Create(task Task) (string, error)  // Returns ID

thinking
**Evaluating code inconsistencies**

I'm noticing that there's a claim semantics check where Assignee is empty, which conflicts with the task loop algorithm filter, indicating some required changes. I’m also considering that MarkInProgress might be redundant. Additionally, it's a good time to check the code in the FilesystemBackend Claim, perhaps especially around line 1494 and later, to see if it has been updated to use flock. There’s quite a bit to unpack here!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1494,1545p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  1494	```go
  1495	// FilesystemBackend stores tasks as JSON files.
  1496	type FilesystemBackend struct {
  1497	    dir string  // e.g., ".gc/tasks/"
  1498	}
  1499	
  1500	func (f *FilesystemBackend) Create(task Task) (string, error) {
  1501	    task.ID = uuid.New().String()[:8]
  1502	    task.CreatedAt = time.Now()
  1503	    task.Version = 1
  1504	    data, _ := json.MarshalIndent(task, "", "  ")
  1505	    return task.ID, os.WriteFile(filepath.Join(f.dir, task.ID+".json"), data, 0644)
  1506	}
  1507	
  1508	func (f *FilesystemBackend) Claim(id string, agent string) (bool, error) {
  1509	    path := filepath.Join(f.dir, id+".json")
  1510	    lockFile := path + ".lock"
  1511	
  1512	    // Advisory lock via flock (automatically released if process crashes)
  1513	    lock, err := os.OpenFile(lockFile, os.O_CREATE|os.O_WRONLY, 0644)

thinking
**Reviewing filesystem changes**

The filesystem backend now utilizes syscall.Flock, which is a good improvement since I initially thought about O_EXCL but we fixed that. Still, I'm wondering if we could face issues with atomic writing and crashes since it's using lockfile and flock. Writing isn't atomic, meaning partial writes might happen in a crash. So, we should implement atomic renames. Also, I’m considering if syscall.Flock is cross-platform; it's suitable for Unix, but what about other environments? We need to adjust our implementation to include ClaimedBy in the claim semantics.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1195,1285p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
  1195	## 6. Event Bus
  1196	
  1197	### 6.1 Event Schema
  1198	
  1199	```go
  1200	type Event struct {
  1201	    Sequence  int64     // Monotonically increasing sequence number
  1202	    Timestamp time.Time
  1203	    Type      EventType
  1204	    Agent     string    // Agent name (empty for system events)
  1205	    Project   string    // Project name (empty for workspace events)
  1206	    Payload   any       // Type-specific payload
  1207	}
  1208	
  1209	type EventType string
  1210	const (
  1211	    EventAgentStarted   EventType = "agent.started"
  1212	    EventAgentStopped   EventType = "agent.stopped"
  1213	    EventAgentStalled   EventType = "agent.stalled"
  1214	    EventAgentCrashed   EventType = "agent.crashed"
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1285,1318p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  1285	
  1286	    // Replay outside lock so publishes are not blocked during replay
  1287	    for _, event := range snapshot {
  1288	        sub.OnEvent(event)
  1289	    }
  1290	}
  1291	
  1292	func (b *EventBus) Unsubscribe(sub EventSubscriber) {
  1293	    b.mu.Lock()
  1294	    defer b.mu.Unlock()
  1295	    // Remove from appropriate list
  1296	    b.critical = removeSubscriber(b.critical, sub)
  1297	    b.optional = removeSubscriber(b.optional, sub)
  1298	}
  1299	```
  1300	
  1301	**Subscriber tiers:**
  1302	- **Critical:** Supervisor, structured logger, hook executor. These block `Publish()` — if they fail, the event is not silently lost.
  1303	- **Optional:** CLI activity feed, metrics aggregator. These receive events asynchronously. Slow consumers may miss events but never stall the bus.
  1304	
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '2140,2185p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  2140	
  2141	---
  2142	
  2143	## 14. Lifecycle Hooks
  2144	
  2145	### 14.1 Hook Execution Model
  2146	
  2147	```go
  2148	type HookExecutor struct {
  2149	    bus      *EventBus
  2150	    hooks    map[EventType][]HookConfig
  2151	    timeout  time.Duration  // Default 30s
  2152	}
  2153	
  2154	func (h *HookExecutor) Run() {
  2155	    events := h.bus.Subscribe(h)  // HookExecutor is a critical subscriber
  2156	    // Events delivered via OnEvent callback
  2157	}
  2158	
  2159	func (h *HookExecutor) OnEvent(event Event) error {

thinking
**Reviewing code updates**

I see they still use `events := h.bus.Subscribe(h)`, but it's returning void, which suggests a mismatch. Maybe there's been a change, but it remains void as of line 1275. Luckily, it’s pseudo, so I can address it. They introduced critical vs. best effort hooks, which is a positive change. In the config schema, they mention environment expansion, so my review needs to consider these updates. I’ll also check the CLI commands for the workflow plan between lines 1835 and 1875.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '1835,1890p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  1835	## 11. CLI Design
  1836	
  1837	### 11.1 Command Structure
  1838	
  1839	The Gas City CLI is `gc`. It extends or replaces `gt` depending on configuration mode.
  1840	
  1841	```
  1842	gc init [--shape ralph|agent-teams|gas-town|custom]
  1843	gc start
  1844	gc stop [--force]
  1845	gc status
  1846	gc level                           # Show capability level and progression
  1847	
  1848	gc agent list
  1849	gc agent start <name>
  1850	gc agent stop <name>
  1851	gc agent attach <name>             # Interactive session (if runtime supports)
  1852	gc agent restart <name> [--graceful]
  1853	gc agent logs <name> [--follow]
  1854	
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '2488,2515p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  2488	---
  2489	
  2490	## 17. Exclusions with Rationale
  2491	
  2492	| Exclusion | Rationale |
  2493	|-----------|-----------|
  2494	| Web dashboard | Important but separate concern. Event bus provides the data API. Dashboard is a consumer, not core SDK. |
  2495	| Cloud/remote execution | v1 is local-only. Docker adapter provides sandboxing. Cloud adapters (EC2, Cloud Run) are Phase 9+. |
  2496	| Multi-tenant isolation | One workspace per user. Multi-tenant requires security boundaries beyond v1 scope. |
  2497	| Agent marketplace | Sharing agent configs/templates is a community feature. The SDK provides the config format; sharing is orthogonal. |
  2498	| Automatic prompt engineering | SDK provides prompt template paths. Generating or optimizing prompts is AI-specific, not orchestration. |
  2499	| Cost management | Responsibility of the underlying runtime (Claude Code has built-in cost tracking). SDK tracks metrics, not billing. |
  2500	| Real-time collaboration | Agents coordinate through tasks and messages, not shared editing. |
  2501	| Hot config reload | Adding agents to a running workspace is desirable but complex. v1 requires restart. |
  2502	
  2503	---
  2504	
  2505	## 18. Formal Properties
  2506	
  2507	### 18.1 Monotonic Capability Growth
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '2555,2595p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  2555	## 19. Open Questions
  2556	
  2557	1. **Out-of-process adapter protocol.** Go plugins are fragile. JSON-RPC over stdin/stdout (LSP-style) is more robust for custom adapters. Decision: implement JSON-RPC in Phase 8.
  2558	
  2559	2. **Config hot-reload.** Can you add agents to a running workspace? v1: no (requires restart). v2: yes (via file watch + diff).
  2560	
  2561	3. **Multi-runtime task serialization.** When results cross runtimes (Claude Code coordinator -> Codex worker), the `TaskResult` schema (Section 2.3) provides the bridge. Adapters normalize their output into this structure. Validate in Phase 4.
  2562	
  2563	4. **State persistence across restarts.** Claude-code uses tmux session survival + `/resume`. Other runtimes need checkpoint files. The `AgentRegistry` persists logical identities to `.gc/agents/` for crash recovery. Define `StateCheckpoint` interface in Phase 3.
  2564	
  2565	5. **`gc` vs `gt` CLI namespace.** Options: (a) `gc` is a separate binary, (b) `gt gc` subcommand, (c) `gt` detects Gas City mode and adjusts behavior. Recommendation: (a) separate binary for clarity.







thinking
**Evaluating the plan details**

Alright, I need to check if the plan includes plugins or hooks. Since this is Round 2, the plan has already integrated some suggestions. I want to quickly evaluate what’s still missing. The plan is quite thorough, but I believe we can propose improvements regarding adoption, tasks claim semantics, and role separation. 

There's a significant issue with task claim semantics, which could be a major bug. Currently, tasks can't be assigned correctly because of how Assignee is defined, so I need to keep that in mind.
thinking
**Identifying algorithm issues**

I'm noticing a problem with the listReady filter, which uses the assignee as self.name. It ends up listing tasks that are already assigned, meaning that if tasks are claimed, they are in progress and therefore not ready. This results in nothing being returned, which is a bug. It seems like they might have meant to use a filter like {assignee: ""}, or the claim semantics should require the assignee to match self.

To fix this, we could propose a robust assignment model and adjust the algorithm to handle tasks that may be either unassigned or assigned.
thinking
**Refining task claiming logic**

My approach allows claiming pre-assigned tasks, but for pool tasks assigned to "devs," I need to ensure that specific instances like "devs[2]" can also claim them. If a pool instance uses the base name "devs," I wonder if I could set the agent parameter to that base name upon claiming. However, tasks might only reference base names, complicating per-instance tracking.

For better clarity, I think storing the `Assignee` as the base name and `WorkerInstance` separately makes sense. Adding a `ClaimedBy` attribute seems like the best implementation. Also, I’ve noticed that the `MarkCompleted` function has a mismatch in expected inputs, which indicates a need for another specification correction.
thinking
**Updating crash recovery expectations**

I believe our changes will enhance both the algorithm and the interface. Now, regarding adoption, the plan mentions that crash recovery uses a persisted agent registry for seamless control without restarting. However, there’s a concern with subprocess adapters because they rely on pipes. After a crash, those pipes are gone, making "resume without restarting" impossible for certain runtimes.

We might need to adjust expectations, limiting adoption to specific runtimes like tmux or Docker while allowing others to restart. I also need to clarify details about the `AgentRegistry` persistence file. Additionally, I’m concerned about optional goroutines possibly leaking in the event bus. Given the user's request for performance and reliability, proposing improvements seems important.
thinking
**Planning proposed changes**

Now, regarding the plugin architecture, I think we should officially make it more central by specifying a plugin directory and manifest, even if it's not strictly required. For context, we need to implement a policy that includes options for keeping, clearing, summarizing, and applying a window, while ensuring it doesn't conflict with automatic prompt engineering.

The `RuntimeAdapter` interface could benefit from streaming support with an `OutputCursor` for large outputs to avoid memory issues. Additionally, I should point out that the GitHub issues backend's claim model causes potential race conditions, making it essential to implement a best-effort strategy for task assignments.

As for crafting revisions, we'll categorize the proposed changes by severity, provide detailed analyses, and create “git-diff style” changes for clarity. It's vital to ensure we align with specifications and reference line numbers in the plan clearly. Each modification will include context lines, focusing on the rationale and the necessity behind each change.
thinking
**Proposing task routing changes**

The user asked for detailed analyses and rationale for each proposed change, along with “git-diff style” modifications reflecting those changes. 

My first focus is on the task routing and claiming semantics. The current setup has the `Task.Assignee` being used not only for routing but also for claims, which isn't working properly. Thus, I propose introducing `ClaimedBy` to improve correctness and allow for greater ease of concurrent task management.

By creating a clearer structure that differentiates between assigned targets and who has claimed the tasks, we can significantly enhance functionality while minimizing disruption. Aside from that, I’m considering adjustments like implementing a filtering mechanism for claimable tasks that could better support various backend systems.
thinking
**Proposing Task Management Changes**

I'm suggesting we improve the task management system by adding a `ClaimableBy` field to the `TaskFilter`. This will let us refine how tasks are routed and claimed. I'll update the algorithm to filter based on the claimable status and adjust methods to count ready tasks according to this new field. 

In the `TaskBackend`, I'll change `CountPending` to `CountReady` and establish clearer expectations within the claiming semantics. Overall, these changes aim to enhance correctness, manage concurrency better, and simplify the process for task claims without introducing excessive complexity.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n \"agent logs|logs\\.jsonl|capture-pane|gc agent logs\" PLAN.md" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
50:3. **Observation lock-in**: Monitoring depends on tmux `capture-pane` output scraping
259:| `claude-code` | tmux session | Yes | `send-keys` / `capture-pane` | Rich observation, requires tmux |
297:| `CaptureOutput()` | `tmux.CapturePane()` | Return piped output from `tmux capture-pane -p` |
1309:| Structured logger | Critical | All events + agent logs | `workspace/.gc/logs.jsonl` (attributed with agent_id) |
1853:gc agent logs <name> [--follow]

thinking
**Exploring log collection possibilities**

I’m realizing there’s no detailed description of how logs are actually gathered, which is a bit of a gap. I think it would be helpful to propose adding something like `EventAgentOutput` to clarify that process. Also, I want to explore the layout of any `.gc/` directories, as that could provide additional insight into how things are structured. Let's dig a bit deeper into these points!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n \"\\.gc/\" PLAN.md | head -n 50" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
110:The agent writes a result file to `<workdir>/.gc/result-<task-id>.json` when done. The adapter watches for this file (via `fsnotify` or polling). The file conforms to the `TaskResult` JSON schema. This is the most reliable mode and the recommended default for custom adapters.
1029:    lockFile     string  // .gc/controller.lock
1030:    socketPath   string  // .gc/controller.sock
1037:2. The controller acquires `.gc/controller.lock` (via `flock`) to prevent multiple controllers per workspace.
1040:5. CLI commands (`gc status`, `gc task list`, etc.) connect to the controller via Unix socket at `.gc/controller.sock`.
1043:**Crash recovery:** If the controller crashes, agents continue running (they are independent OS processes). On next `gc start`, the controller rediscovers running agents via the persisted agent registry (`.gc/agents/*.json`) and resumes control without restarting them.
1309:| Structured logger | Critical | All events + agent logs | `workspace/.gc/logs.jsonl` (attributed with agent_id) |
1396:**Progress reporting:** Agents can update `LastProgressUpdate` and `ProgressMessage` by writing to a well-known progress file (`<workdir>/.gc/progress.json`) or through a heartbeat mechanism. The runtime adapter's `GetState()` reads this file to populate the fields. This distinguishes "slow legitimate work" (frequent progress updates) from "actual stall" (no updates for `stuck_threshold`).
1492:**Filesystem backend:** JSON files in `.gc/tasks/`. Zero dependencies. Good for simple setups and testing. `Claim()` uses filesystem locking (`flock`).
1497:    dir string  // e.g., ".gc/tasks/"
2098:    stateDir   string                   // .gc/agents/ for persistence
2107:    r.persist(handle)  // Write to .gc/agents/<addr>.json for crash recovery
2563:4. **State persistence across restarts.** Claude-code uses tmux session survival + `/resume`. Other runtimes need checkpoint files. The `AgentRegistry` persists logical identities to `.gc/agents/` for crash recovery. Define `StateCheckpoint` interface in Phase 3.
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "completion_mode" PLAN.md' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
106:completion_mode = "prompt"    # "file", "prompt", or "exit"


thinking
**Addressing configuration inconsistencies**

I’m noticing some conflicts with the TaskStatus type used in TaskResult and its earlier definition in the tasks section. This might conflict with AgentStatus too. There are comments in TaskResult about its status but no clarity, especially in the runtime adapter. We really need to unify these spec inconsistencies. It's crucial to propose a change to normalize config types for consistency. This will help implementers avoid missteps and ensure everything aligns properly with the schema.
thinking
**Proposing a spec hygiene change**

I think proposing a hygiene change for the spec is essential for ensuring accuracy. It might be worth examining how AgentConfig is being used, especially since there could be a resolved config struct in the future. I’ll look for the definition of `WorkspaceConfig` to see if that helps clarify things. Maintaining correctness in our specifications is crucial for smooth implementation, so I want to make sure we're on the right track here!
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "type WorkspaceConfig|type AgentEntry|AgentEntry" PLAN.md | head -n 50' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
1978:    cfg.Agents = append(cfg.Agents, AgentEntry{Name: "coordinator", Role: "coordinator", Scope: "workspace"})
1979:    cfg.Agents = append(cfg.Agents, AgentEntry{Name: "supervisor", Role: "supervisor", Scope: "workspace"})
1981:        cfg.Agents = append(cfg.Agents, AgentEntry{Name: "observer-" + rig.Name, Role: "observer", Scope: "project"})
1982:        cfg.Agents = append(cfg.Agents, AgentEntry{Name: "integrator-" + rig.Name, Role: "integrator", Scope: "project"})
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "type WorkspaceConfig" PLAN.md | head -n 50' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:










thinking
**Evaluating task statuses and updates**

I'm looking at TaskResult.Status and noticing that it uses the TaskStatus type, which is mentioned in the tasks section. It seems like separating runtime-level status from task system status might be a good idea for clarity. I think I should propose changes to make the specification more consistent—like including various configurations in AgentConfig. The user asked about architecture, so these updates could be captured as Change 2. Also, I realize we need to check on auto-detection for tmux permissions.
thinking
**Reviewing code and specs**

I’m noticing some inconsistencies in the built-in runtime adapters list, like how "codex" is mentioned as a SubprocessAdapter with DefaultCommand, yet the earlier struct lacks this. It feels important to focus on meaningful improvements in my review rather than being overly picky. I need to identify the final proposed changes that maximize value. I've found several critical issues, such as the task claim semantics and spec inconsistencies regarding AgentConfig. I should also address the plugin architecture and see if there’s any mention of a plugin directory in the plan.
thinking
**Creating output diffs**

I need to focus on producing output only, without any tool outputs or file modifications—got it! I should create each change with a diff format, showing “git-diff style” changes relative to the original plan. Since the user asked for individual changes, it might be best to separate them into distinct diff blocks for clarity. This way, each diff will have its own block specified as `diff --git a/PLAN.md b/PLAN.md`. Now, I’ll start crafting the diff for Change 1, including modifications to sections 4.1 and 8.
thinking
**Proposing algorithm changes**

I'm looking at the existing algorithm and proposing several updates. First, I want to include both unassigned tasks and those assigned to the base agent or pool when listing claimable tasks. Instead of using the agent's name, I should claim tasks using the identity string. I'll also update the MarkCompleted function to reflect these changes.

Additionally, in the TaskBackend interface, I need to adjust the List filter to include ClaimableBy and update the Task struct to include ClaimedBy. I'll also be sure to reflect these changes in the relevant sections of the algorithm. Time to craft the diff!
thinking
**Outlining significant changes**

I’m updating the claim semantics, ensuring that tasks must be ready and not claimed by anyone when checked. I also need to define the `matchesAssignee` function. We should reflect these updates in the FilesystemBackend Claim to keep everything current.

Next, I’ll modify the RuntimeAdapter interface to add an optional Adoption interface for crash recovery, and I'll update the related paragraphs. There's also a push to align the config schema with key elements like completion mode, health, pool, and other details, making the structure more cohesive. Lastly, I’ll plan out several major changes, including improving the event bus and defining a batch tracking system. Time to craft some well-structured diffs!
thinking
**Drafting the review analysis**

I’m starting the review by laying out the proposed changes clearly, focusing on providing detailed analysis and justification for each one. My approach ensures that I adhere to the instruction to include a git-diff style change with relevant snippets.

For the first change, I’ll revise the task claim semantics and algorithm in several sections. This update requires context around the code to show what's being adjusted. I'm also looking to clean up the TaskBackend interface and related functions to enhance clarity and functionality.

Then, I’ll tackle the crash recovery adoption, modifying the necessary interfaces and updating associated sections. Aligning the configuration schema with completion modes is another key task. Finally, I’ll ensure that my diffs are well-documented, combining clarity with precision to make the changes easily understandable.
thinking
**Updating configuration and workflows**

I need to confirm the location of the contract tests in section 15.2, where we reference `cfg.Health.PingTimeout`. I’ll also revisit the `AgentConfig` struct, which lacks some essential fields—I'll add those definitions for nested config types to ensure consistency. 

For the next proposed changes, I’ll add a `startup_priority` field to the agent schema and adjust the validation rules to decouple from roles. This will enhance flexibility. 

Lastly, I'll propose a new section for batch processing, ensuring it aligns with existing structures while introducing commands for batch management. Step by step, I want to make clear, thoughtful updates!
thinking
**Proposing structural enhancements**

I'll be adding a new section related to event bus scaling and durable event logs, proposing to introduce `agent.output` as a new event type. For the EventBus implementation, I'm suggesting updating it to utilize a bounded queue for subscribers instead of using multiple goroutines, which could help reduce lag.

Next, I want to formalize the plugin system by defining a protocol for external adapters found in a specified directory. This should support extensibility, especially for integrating custom runtimes.

Lastly, I'll review the context management policy to enhance session handling, possibly replacing `clear_context` with a `context_policy` string for better configurability. I'll also address necessary changes while keeping user experience in mind!
thinking
**Proposing secret redaction improvements**

I believe we should include a proposal for secret redaction as a separate change, possibly Change 9. While it could fit into Change 3 or Change 7, I think it's best to keep it distinct to emphasize its importance. This change would focus on config layering and include details about environment variable expansion and secret redaction. 

I'm ready to craft the review, ensuring I include all necessary components and line references for clarity and credibility, sticking to the formatting rules. Let's make it clear and concise!
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "''^'"\\[batches\\]\" -n PLAN.md" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
576:[batches]
863:[batches]
exec
/bin/zsh -lc 'cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n "no hardcoded role logic" PLAN.md' in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
950:There is no hardcoded role logic in the SDK. The SDK provides the infrastructure (task dispatch, health monitoring, messaging, etc.) and the role definition provides the policy (what the agent does, when, and how).
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '165,210p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
   165	
   166	// AgentHandle is a transient reference to a running agent instance.
   167	// A new handle is created on each Start(). The Identity persists.
   168	type AgentHandle struct {
   169	    ID        string            // Unique ID for this runtime instance (UUID)
   170	    Identity  AgentIdentity     // Logical identity (stable across restarts)
   171	    Runtime   string            // Runtime adapter name (e.g., "claude-code")
   172	    StartedAt time.Time
   173	    PID       int               // OS process ID (0 if not applicable)
   174	    Metadata  map[string]string // Runtime-specific metadata
   175	}
   176	
   177	// AgentConfig is the resolved config for starting an agent.
   178	type AgentConfig struct {
   179	    Name          string
   180	    Role          string
   181	    Runtime       string
   182	    Project       string            // Empty for workspace scope
   183	    Command       string            // Primary command (e.g., "claude")
   184	    Args          []string
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '485,520p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
   485	theme = "default"               # "default" or "gas-town" (enables GT naming in output)
   486	
   487	# === LEVEL 0: Agent Runtime ===
   488	[[agents]]
   489	name = "string"                 # Required. Agent identifier (unique within workspace).
   490	runtime = "string"              # Runtime adapter: "claude-code", "codex", "agent-sdk",
   491	                                # "subprocess", "docker"
   492	                                # If omitted, auto-detected:
   493	                                #   1. Check for `claude` binary -> "claude-code"
   494	                                #   2. Check for `codex` binary -> "codex"
   495	                                #   3. Check for `gemini` binary -> "gemini-cli"
   496	                                #   4. Check for `docker` binary -> "docker"
   497	                                #   5. Fallback -> "subprocess"
   498	role = "worker"                 # "coordinator", "supervisor", "observer", "integrator",
   499	                                # "worker", "agent", "service", or any custom string
   500	scope = "project"               # "workspace" (one total) or "project" (one per project)
   501	ephemeral = false               # Whether this agent is created/destroyed per-task
   502	isolation = "worktree"          # Isolation strategy:
   503	                                # "none"      - Runs in workspace root (simple, shared state)
   504	                                # "worktree"  - Git worktree (Gas Town default, branch isolation)
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n \"### 2\\.8\" PLAN.md" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
402:### 2.8 Adapter Registry
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '388,460p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
   388	            envMap[parts[0]] = parts[1]
   389	        }
   390	    }
   391	    for k, v := range configEnv {
   392	        envMap[k] = v
   393	    }
   394	    result := make([]string, 0, len(envMap))
   395	    for k, v := range envMap {
   396	        result = append(result, k+"="+v)
   397	    }
   398	    return result
   399	}
   400	```
   401	
   402	### 2.8 Adapter Registry
   403	
   404	```go
   405	var (
   406	    adapterMu       sync.RWMutex
   407	    adapterRegistry = map[string]RuntimeAdapter{

thinking
**Modifying implementation phases**

I’m adjusting the section for the plugin loader. In the implementation phases, particularly Phase 8, I need to include "Additional Adapters + Conformance CLI." I’ll make those updates. Also, I want to check the line numbers around this phase, specifically around 2388. I'll take a look at lines 2388 to 2455 to ensure everything aligns correctly. Let’s make sure these changes fit nicely within the context!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '2388,2470p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 54ms:
  2388	## 16. Implementation Phases
  2389	
  2390	### Phase 1: Agent Runtime Extraction (2-3 weeks)
  2391	
  2392	**Goal:** Extract the `claude-code` adapter from Gas Town's existing tmux code without changing behavior.
  2393	
  2394	**Deliverables:**
  2395	- `internal/runtime/adapter.go` -- RuntimeAdapter interface + ReadinessProbe
  2396	- `internal/runtime/registry.go` -- Thread-safe adapter registry
  2397	- `internal/runtime/types.go` -- AgentHandle, AgentIdentity, AgentConfig, TaskResult
  2398	- `internal/runtime/claude_code/` -- Claude Code adapter wrapping `internal/tmux/`
  2399	- `internal/runtime/subprocess/` -- Generic subprocess adapter
  2400	- Contract test suite in `internal/runtime/contract_test.go`
  2401	- `gc agent start/stop/status` commands
  2402	
  2403	**Acceptance:** All existing Gas Town tests pass. Contract tests (happy + error paths) pass for both adapters.
  2404	
  2405	### Phase 2: Config Parser + CLI Foundation (1-2 weeks)
  2406	
  2407	**Goal:** Parse `gas-city.toml`, detect levels, validate, and provide `gc init` + `gc level` + `gc validate`.
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '2470,2495p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 52ms:
  2470	
  2471	**Deliverables:**
  2472	- `internal/migration/` -- Gas Town -> Gas City migration with MigrationReport
  2473	- `gc migrate` command with `--dry-run` and structured report output
  2474	- Level 7 config with all role behaviors and DAG dependencies
  2475	- Observer, integrator, service role behaviors
  2476	
  2477	**Acceptance:** `gc migrate --dry-run` generates correct config + report from Gas Town workspace. Report surfaces hooks/env-vars needing manual review. Gas Town shape starts all agents with DAG ordering.
  2478	
  2479	### Phase 8: Additional Adapters + Conformance CLI (ongoing)
  2480	
  2481	**Goal:** Codex, Gemini, Agent SDK, Docker adapters + `gc test-adapter`.
  2482	
  2483	**Deliverables:**
  2484	- Per-adapter package in `internal/runtime/`
  2485	- Contract tests passing for each (happy + error paths)
  2486	- `gc test-adapter <name>` CLI command for ecosystem certification
  2487	
  2488	---
  2489	

thinking
**Updating plugin phases**

I'm considering whether to update the phases for the plugin system, perhaps moving to Phase 2 or Phase 8. It seems like a good idea to propose a new deliverable for that. Now, I think the config layering could fit into section 3.1. I should definitely open up to lines 463-486 to gather more details. Let's make sure we have all the necessary information to proceed!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && nl -ba PLAN.md | sed -n '463,490p'" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
   463	## 3. Configuration Schema
   464	
   465	### 3.1 File Format and Location
   466	
   467	- **Format:** TOML (consistent with Gas Town's formula system)
   468	- **File:** `gas-city.toml` at the workspace root
   469	- **Detection logic:**
   470	  1. `gas-city.toml` exists -> Gas City mode
   471	  2. `mayor/town.json` exists without `gas-city.toml` -> Gas Town compatibility mode
   472	  3. Neither exists -> fresh workspace, `gc init` required
   473	
   474	- **Environment variable expansion:** Values in `env` maps support `${VAR}` expansion from the host environment and optional `.env` files in the workspace root. API keys and secrets should use this mechanism rather than being hardcoded in `gas-city.toml`.
   475	
   476	- **TOML ordering note:** Sub-tables like `[agents.pool]` and `[agents.health]` attach to the most recently declared `[[agents]]` entry. Each agent's sub-tables must appear immediately after its `[[agents]]` header and before the next `[[agents]]` header. The config validator detects and reports orphaned sub-tables.
   477	
   478	### 3.2 Full Schema
   479	
   480	```toml
   481	# === REQUIRED ===
   482	[workspace]

thinking
**Evaluating plugin integration**

Section 3.1 already includes environment expansion and `.env` file support, which is great since that means we don't need to propose anything there. However, we should still consider suggesting redaction and config layering. I need to check if the plan mentions the plugin directory, specifically looking for `.gc/plugins` or the term "plugin." I want to make sure everything's covered properly!
exec
/bin/zsh -lc "cd /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge && rg -n \"plugins?|\\.gc/plugins\" PLAN.md | head -n 50" in /Users/csells/Code/Forks/steveyegge/gastown/gas-city-sdk-spec-forge succeeded in 51ms:
28:- Extensibility: sandboxes (isolation modes), plugins (adapter registry), hooks (lifecycle events)
2557:1. **Out-of-process adapter protocol.** Go plugins are fragile. JSON-RPC over stdin/stdout (LSP-style) is more robust for custom adapters. Decision: implement JSON-RPC in Phase 8.




thinking
**Identifying changes needed**

I'm analyzing some technical aspects I came across. There's an issue with completion_mode at line 106, but schema is missing it, and AgentConfig doesn't have it either. 

I've identified several components needing updates, like missing batches specs, unbounded goroutines in the event bus, and a lack of context policy. 

I think I've got around eight essential changes to address. It all revolves around adjusting claims, adding support for adoption, aligning schemas, and formalizing the plugin system.
thinking
**Clarifying diff strategy**

I'm thinking about adding a small "nits" section at the end and referencing gas-city-vision.md along with mapping. 

Next, I should create diffs for each change, ensuring they contain multiple hunks as needed. It’s important not to create conflicts, even if some changes modify the same lines—it’s okay since each diff is conceptual. 

The user wants diffs versus the original plan, which could lead to some ambiguity. I’m considering treating the diffs as incremental changes rather than cumulative or independent patches. A combined diff at the end, along with individual diffs, might be clearer.
thinking
**Crafting focused diffs**

The user has asked for diffs related to each proposed change. I’ll ensure that each diff focuses on specific hunks, and if modifications overlap, we need to handle those carefully. 

For example, change 3 affects the config schema with runtime_config while change 4 adds startup_priority and monitoring; they impact the same area but can still be separate hunks. 

If the user tries to apply change 4 without 3, that’s alright since the overlap is minimal. I’ll also ensure that I use ASCII for the diffs and avoid em dashes in the text. Now, I’ll start crafting each change.
thinking
**Crafting detailed review feedback**

First, I need to make sure I respect the instruction that my output is strictly my review feedback. Before I start writing, I might want to check for other spec documents in the parent directory, but since the user specifically requested the vision, that’s okay.

I’ll begin with bullet mapping from the `../specs/gas-city-vision.md`, mentioning each bullet and its alignment with the plan. I’ll also propose changes, pointing out any mismatches with the vision, such as the plugin architecture and roles not being fully external. Given that the user wants detail, I can prepare eight changes with a couple of paragraphs each along with diffs. It’s going to be lengthy, but it should be fine to go ahead!
codex
**Vision Check (`../specs/gas-city-vision.md`)**  
The plan broadly satisfies the vision bullets: toolkit/not-one-shape (`PLAN.md:13-29`, `PLAN.md:643-1009`), progressive levels (`PLAN.md:581-626`), user-defined roles via external prompts (`PLAN.md:923-1009`), sandboxes/isolation (`PLAN.md:501-506`), uniform “factory worker” abstraction (`PLAN.md:56-88`). The remaining *vision gap* is “plugins” as a first-class extensibility surface: today it is “adapter registry + open question” (`PLAN.md:28`, `PLAN.md:402-432`, `PLAN.md:2557-2564`), not an actually usable plugin model.

Below are the biggest Round 2 improvements still missing, ordered by severity. Each includes rationale and a `git diff` style patch against the current `PLAN.md`.

---

## 1) Fix Task Claiming vs Assignment (Currently Inconsistent, Breaks Ralph + Pools)

**Problem**  
The Ralph loop lists “ready tasks assigned to me” (`PLAN.md:675`), but `Claim()` only succeeds when `Assignee == ""` and then sets `Assignee = agent` (`PLAN.md:1484`). That makes the loop logically impossible: tasks “assigned to me” are non-claimable by definition. Same issue for pools: tasks pre-assigned to a pool name (which is implied by `CountPending(assignee)` usage) can never be claimed under the current semantics. There is also a type mismatch: `MarkCompleted` expects a `TaskResult` but the loop passes `result.Output` (`PLAN.md:687` vs `PLAN.md:1432`).

**Best fix**  
Split “routing target” from “exclusive claim lock.” Keep `Assignee` as the *logical target* (agent name or pool name). Add `ClaimedBy` as the *actual* claimant identity (including pool instance). Change `Claim()` to set `ClaimedBy` and `Status`, not overwrite routing. Add a `TaskFilter.ClaimableBy` so loops can ask “what can I claim?” without encoding pool-matching rules in every caller.

**Why it’s better**  
This makes Ralph, Agent Teams, and “coordinator assigns work to a pool” all coherent. It also preserves important information: “this task was routed to pool devs” and “this specific worker instance claimed it,” which matters for debugging, retries, and for implementing fair scheduling.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -672,21 +672,21 @@
 LOOP:
-  tasks = taskBackend.ListReady(filter: {assignee: self.name, sortBy: "priority", sortOrder: "desc"})
+  tasks = taskBackend.List(filter: {status: Ready, claimable_by: self.identity, sortBy: "priority", sortOrder: "desc"})
   if len(tasks) == 0:
     sleep(poll_interval)
     goto LOOP
   task = tasks[0]  // highest priority
-  ok = taskBackend.Claim(task.id, self.name)  // Atomic claim
+  ok = taskBackend.Claim(task.id, self.identity)  // Atomic claim (records ClaimedBy)
   if !ok:
     goto LOOP  // Another agent claimed it first
   runtime.Assign(self.handle, task)
   WAIT_COMPLETION:
     result = runtime.WaitForResult(self.handle)  // Blocks with internal timeout
     if result.Status == Completed:
-      taskBackend.MarkCompleted(task.id, self.name, result.Output)
+      taskBackend.MarkCompleted(task.id, self.identity, result)
       if clear_context:
         runtime.Stop(self.handle, graceful=true)
         self.handle = runtime.Start(ctx, self.config)
       goto LOOP
     if result.Status == Failed:
       taskBackend.MarkFailed(task.id, self.name, result.Error)
       runtime.Restart(self.handle)
       goto LOOP

@@ -1421,34 +1421,43 @@
 type TaskBackend interface {
     // CRUD
     List(filter TaskFilter) ([]Task, error)
     Get(id string) (Task, error)
     Create(task Task) (string, error)  // Returns ID
     Update(id string, updates TaskUpdates) error
     CountPending(assignee string) (int, error)

     // Atomic operations for concurrency safety
-    Claim(id string, agent string) (bool, error)                     // Atomic: claim or return false
+    Claim(id string, claimant string) (bool, error)                  // Atomic: claim or return false
     MarkInProgress(id string, agent string) error
     MarkCompleted(id string, agent string, result TaskResult) error
     MarkFailed(id string, agent string, reason string) error
 }

 type Task struct {
@@
-    Assignee     string      // Agent name
+    Assignee     string      // Logical target: agent name or pool name (routing). Empty means unassigned/pullable.
+    ClaimedBy    string      // Claimant identity (e.g. "ws/proj/devs[2]"). Empty means unclaimed.
@@
 }

 type TaskFilter struct {
     Status    TaskStatus
     Assignee  string
+    ClaimableBy string // If set, backend returns only tasks this claimant is eligible to Claim()
     Project   string
     Labels    []string
     NoDeps    bool    // Only tasks with no unresolved dependencies
     SortBy    string  // "priority", "created", "updated"
     SortOrder string  // "asc", "desc"
 }

@@ -1484,7 +1493,20 @@
-**Claim semantics:** `Claim(id, agent)` atomically checks `Status == Ready && Assignee == ""`, sets `Status = InProgress, Assignee = agent`, and increments `Version`. If the task is already claimed, it returns `(false, nil)`. This prevents two agents from claiming the same task.
+**Claim semantics:** `Claim(id, claimant)` atomically checks:
+- `Status == Ready`
+- `ClaimedBy == ""`
+- AND the claimant is eligible:
+  - If `Assignee == ""`, anyone can claim (pull model).
+  - If `Assignee != ""`, the claimant must match the assignee (push model), including pool instances:
+    - Example: `Assignee = "devs"` is claimable by `"ws/proj/devs[0]"`, `"ws/proj/devs[1]"`, etc.
+
+On success it sets `Status = InProgress`, sets `ClaimedBy = claimant`, increments `Version`,
+and leaves `Assignee` unchanged (preserves routing). If already claimed, returns `(false, nil)`.

@@ -1508,7 +1530,7 @@
-    if task.Status != StatusReady || task.Assignee != "" {
+    if task.Status != StatusReady || task.ClaimedBy != "" {
         return false, nil  // Already claimed
     }
-    task.Status = StatusInProgress
-    task.Assignee = agent
+    task.Status = StatusInProgress
+    task.ClaimedBy = agent
     task.Version++
     task.UpdatedAt = time.Now()
```

---

## 2) Make Crash Recovery Realistic: “Adopt If Possible, Otherwise Restart”

**Problem**  
The plan promises “controller rediscovers running agents … and resumes control without restarting them” (`PLAN.md:1043`). That is not generally possible for subprocess-based adapters: after the controller dies, you cannot recover stdin/stdout pipes. Even for tmux you need an explicit “adopt/attach to existing runtime instance” operation that is not in the adapter contract (`PLAN.md:56-88`).

**Best fix**  
Introduce an *optional* adoption interface. The controller uses it when available (tmux, docker). If unsupported, the controller explicitly restarts the agent (and records that in the registry/events).

**Why it’s better**  
This makes the promise true for adapters that can support it, and avoids “spec says X but implementation cannot” for ones that cannot. It also clarifies which state you need to persist in `AgentHandle.Metadata` to make adoption work.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -90,6 +90,16 @@
 type ReadinessProbe interface {
     IsReady(handle AgentHandle) (bool, error)
     WaitReady(ctx context.Context, handle AgentHandle) error
 }
+
+// Adoption is an optional interface adapters can implement to support
+// controller crash recovery without forcing an agent restart.
+type Adoption interface {
+    // Adopt attaches to an already-running runtime instance described by the
+    // persisted handle (PID/session/container ID in handle.Metadata).
+    // Returns an updated handle or ErrNotSupported.
+    Adopt(ctx context.Context, persisted AgentHandle) (AgentHandle, error)
+}

@@ -1043,7 +1053,13 @@
-**Crash recovery:** If the controller crashes, agents continue running (they are independent OS processes). On next `gc start`, the controller rediscovers running agents via the persisted agent registry (`.gc/agents/*.json`) and resumes control without restarting them.
+**Crash recovery:** If the controller crashes, agents may continue running (they are independent OS processes). On next `gc start`, the controller loads persisted handles from `.gc/agents/*.json` and:
+1. If the adapter implements `Adoption`, it calls `Adopt()` to re-attach without restart (tmux, docker).
+2. Otherwise it restarts the agent with `Start()` and publishes an event noting the restart (subprocess-style runtimes cannot recover I/O pipes).
```

---

## 3) Make Spec Internals Consistent: `completion_mode` + AgentConfig Fields Must Match Usage

**Problem**  
The spec defines `completion_mode` in `[agents.runtime_config]` (`PLAN.md:102-117`), but it is missing from the “Full Schema” (`PLAN.md:509-514`). Separately, `AgentConfig` is missing fields that the plan later treats as present (notably health timeouts in contract tests and pool readiness usage).

**Best fix**  
Add the common `completion_mode` knobs to the schema, and update `AgentConfig` in Section 2.3 to include the fields the rest of the spec assumes exist (health/loop/pool/prompts/hooks and completion config). Keep `RuntimeConfig map[string]any` for adapter-specific extras.

**Why it’s better**  
This eliminates “paper cuts” where implementers build the wrong structs or omit key fields, and it makes the contract tests and controller algorithms materially implementable.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -177,6 +177,30 @@
 type AgentConfig struct {
     Name          string
     Role          string
     Runtime       string
     Project       string            // Empty for workspace scope
+    Scope         string            // "workspace" or "project"
+    Ephemeral     bool
+    Isolation     string            // "none"|"worktree"|"directory"|"container"
+
     Command       string            // Primary command (e.g., "claude")
     Args          []string
     Env           map[string]string
     WorkDir       string
     SystemPrompt  string            // Path or inline prompt
     NudgePrompt   string            // Initial message to send
+
+    // Completion detection (Section 2.2.1)
+    CompletionMode     string // "file"|"prompt"|"exit"
+    PromptRegex        string // used by prompt mode (optional)
+    ResultFileTemplate string // used by file mode (optional)
+
+    // Orchestration policies (schema sections)
+    Health   AgentHealthConfig
+    Pool     AgentPoolConfig
+    Loop     AgentLoopConfig
+    Prompts  AgentPromptConfig
+    Hooks    AgentHookConfig
+
     RuntimeConfig map[string]any    // Adapter-specific settings
     DependsOn     []string          // Agent names that must start first
 }

@@ -509,6 +533,12 @@
 [agents.runtime_config]         # Adapter-specific settings
 command = "claude"              # Command to run
 args = ["--dangerously-skip-permissions"]
 env = { KEY = "${MY_SECRET}" }  # Extra env vars (supports ${VAR} expansion)
+completion_mode = "prompt"      # "file", "prompt", or "exit" (Section 2.2.1)
+prompt_regex = ""              # Optional: prompt detection pattern for "prompt" mode
+result_file_template = ""       # Optional: e.g. ".gc/result-{{.TaskID}}.json" for "file" mode
 resume_flag = "--resume"        # How this runtime resumes sessions
 resume_style = "flag"           # "flag" or "subcommand"
```

---

## 4) Remove Hidden Role Coupling: Make Monitoring + Startup Ordering Explicit (Toolkit, Not “Blessed Roles”)

**Problem**  
The plan claims “no hardcoded role logic” (`PLAN.md:950`), but it still hardcodes role-based priorities (`PLAN.md:1051-1062`), makes Level 6 implicit on “supervisor role exists” (`PLAN.md:571-573`, `PLAN.md:613-615`), and enforces role-count limits (`PLAN.md:636-637`). That is fine for a product, but it is at odds with the “build any town shape” toolkit posture.

**Best fix**  
Add explicit knobs:
- `startup_priority` (per agent) so startup ordering is not coupled to role strings.
- `[monitoring]` (Level 6) so monitoring is a configured capability, not inferred from a magic role.
- Relax “only one coordinator/supervisor” from hard errors to either warnings or an optional “strict mode.”

**Why it’s better**  
This preserves progressive levels while removing “reserved role” behavior from the core. It makes the SDK more general and less surprising for Level 8 users building non-Gas-Town topologies.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -498,6 +498,8 @@
 role = "worker"                 # "coordinator", "supervisor", "observer", "integrator",
                                 # "worker", "agent", "service", or any custom string
+startup_priority = 2            # Optional. Lower starts earlier when depends_on is empty.
 scope = "project"               # "workspace" (one total) or "project" (one per project)
 ephemeral = false               # Whether this agent is created/destroyed per-task

@@ -571,9 +573,11 @@
-# === LEVEL 6: Monitoring (implicit when supervisor role exists) ===
-# No extra config section needed — monitoring is enabled when a
-# supervisor agent is configured.
+# === LEVEL 6: Monitoring ===
+[monitoring]
+enabled = false
+patrol_interval = "30s"

@@ -613,8 +617,8 @@
-    if level >= 5 && hasSupervisor(cfg) {
+    if level >= 5 && cfg.Monitoring != nil && cfg.Monitoring.Enabled {
         level = 6 // Level 6: Monitoring
     }

@@ -636,8 +640,8 @@
-| Coordinator limit | At most 1 coordinator per workspace | "multiple coordinators not supported" |
-| Supervisor limit | At most 1 supervisor per workspace | "multiple supervisors not supported" |
+| Coordinator limit | (warning) Multiple coordinators allowed; `gc init` uses 1 by default | "warning: multiple coordinators configured" |
+| Supervisor limit | (warning) Multiple monitors allowed; `gc init` uses 1 by default | "warning: multiple supervisors configured" |

@@ -1049,19 +1053,13 @@
-Agents start according to a dependency DAG. If no `depends_on` is specified, agents use default priority groups based on role.
+Agents start according to a dependency DAG. If no `depends_on` is specified, agents use `startup_priority` (default 2).

-// Default priorities (used when depends_on is not specified)
-var defaultPriority = map[string]int{
-    "supervisor":  0,
-    "coordinator": 1,
-    ...
-}
-// Custom roles default to priority 2 (same as per-project agents)
+// Startup priority is configured per agent. `gc init` chooses sensible defaults for known shapes.
```

---

## 5) Either Specify Batches (Convoys) or Remove Them From Schema (Dangling Feature Right Now)

**Problem**  
`[batches]` is present in the schema and the Gas Town example (`PLAN.md:575-578`, `PLAN.md:863-866`), but there is no batch data model, no persistence, no events, and no CLI. As written, it is an unimplemented promise and undermines “spec is implementable.”

**Best fix**  
Define batches as a controller-side, backend-agnostic grouping stored in `.gc/batches/` (so it works with GitHub issues, beads, filesystem tasks). Make `auto_track` create a batch whenever a workflow dispatches multiple tasks.

**Why it’s better**  
This gives users a compelling “orchestration view” without forcing every TaskBackend to grow a batch concept. It also helps with multi-project operations (the whole point of Level 7).

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -1540,6 +1540,62 @@
 ---
 
 ## 9. Messaging System
+
+@@
+## 8. Task System
+...
+### 8.3 Batch (Convoy) Tracking
+
+Batches are a workspace-local grouping of tasks for tracking multi-task dispatches (Gas Town "convoys").
+Batches are stored in `.gc/batches/<batch-id>.json` and reference task IDs in the configured TaskBackend.
+This keeps batch tracking backend-agnostic (works with beads, GitHub Issues, filesystem).
+
+```go
+type BatchStatus string
+const (
+    BatchRunning   BatchStatus = "running"
+    BatchCompleted BatchStatus = "completed"
+    BatchFailed    BatchStatus = "failed"
+)
+
+type Batch struct {
+    ID          string
+    Title       string
+    Description string
+    Project     string   // optional
+    TaskIDs     []string
+    Status      BatchStatus
+    CreatedAt   time.Time
+    UpdatedAt   time.Time
+}
+```
+
+**auto_track behavior (`[batches] auto_track = true`):**
+1. When a workflow dispatches 2+ tasks, create a new Batch referencing those task IDs.
+2. Batch status is derived from task statuses:
+   - Completed if all tasks Completed
+   - Failed if any task Failed (configurable later)
+   - Running otherwise
+
+**CLI additions:**
+`gc batch list`, `gc batch show <id>`, `gc batch watch <id>`
```

(If you do *not* want to ship batches in v1, the better alternative is to remove `[batches]` from the schema and Gas Town example until specified.)

---

## 6) Make the Event Bus Safe Under Load + Define How “Agent Logs” Become Events

**Problem**  
The bus still spawns one goroutine per optional subscriber per event (`PLAN.md:1267-1270`). Under bursty output (agent logs, workflow step spam), this becomes unbounded goroutines and can destabilize the controller. Separately, the plan claims “structured logger subscribes to all events + agent logs” (`PLAN.md:1309`) but there is no event type or ingestion path for agent output.

**Best fix**  
1. Replace “goroutine per event” with per-subscriber bounded queues and a single worker goroutine per subscriber (preserves ordering, bounds memory/cpu).  
2. Introduce `agent.output` events and a controller “log pump” that reads from `RuntimeAdapter.CaptureOutput()` (or polls `ReadOutput()` for snapshot adapters) and publishes normalized log line events.

**Why it’s better**  
This is the difference between “works in a demo” and “can run overnight with 10 agents.” It also makes `gc agent logs --follow` and `gc activity --follow` implementable without adapter-specific hacks.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -1210,6 +1210,7 @@
 const (
@@
     EventAgentReady     EventType = "agent.ready"
+    EventAgentOutput    EventType = "agent.output" // normalized log lines/frames
@@
 )

@@ -1238,12 +1239,19 @@
 type EventBus struct {
     mu          sync.RWMutex
     sequence    int64
-    critical    []EventSubscriber  // Block on these (supervisor, structured logger)
-    optional    []EventSubscriber  // Fire-and-forget (CLI feed, metrics)
+    critical    []*subscriberState // bounded queue + single worker per subscriber
+    optional    []*subscriberState // bounded queue + single worker per subscriber
     buffer      *ring.Buffer[Event]  // Ring buffer (10k events) for catch-up on subscribe
 }
+
+type subscriberState struct {
+    sub   EventSubscriber
+    ch    chan Event
+    drop  bool // true for optional subscribers: drop when full
+}

@@ -1267,10 +1275,11 @@
-    // Optional subscribers: fire-and-forget
-    for _, sub := range optionals {
-        go sub.OnEvent(event)
-    }
+    // Optional subscribers: enqueue (drop when full), no unbounded goroutines
+    for _, st := range optionals {
+        select { case st.ch <- event: default: /* drop */ }
+    }

@@ -1308,7 +1318,8 @@
-| Structured logger | Critical | All events + agent logs | `workspace/.gc/logs.jsonl` (attributed with agent_id) |
+| Structured logger | Critical | All events + agent.output | `workspace/.gc/logs.jsonl` (attributed with agent_id) |
 | Hook executor | Critical | Configurable per event type | Runs user-defined shell commands |
```

---

## 7) Turn “Plugins” From an Open Question Into a Usable Extensibility Surface

**Problem**  
Right now, adding a runtime adapter requires modifying Go and recompiling (`PLAN.md:402-432`). The vision explicitly calls out plugins (`../specs/gas-city-vision.md:24-25`). The plan mentions JSON-RPC only as a future “Phase 8” open question (`PLAN.md:2557`, `PLAN.md:2479-2487`), which delays the most “toolkit-y” part of the toolkit.

**Best fix**  
Define the plugin boundary now (protocol + discovery), even if only a minimal loader ships early:
- Plugin discovery: workspace-local `.gc/plugins/` plus user-global `~/.gc/plugins/`.
- Protocol: JSON-RPC over stdio with version handshake and capability advertisement.
- Support at least runtime adapters; ideally also task/message backends later.

**Why it’s better**  
This is what lets Level 8 users wire in internal models, bespoke sandboxes, or custom execution substrates without forking Gas City. It also reduces long-term architectural risk: you avoid baking in an internal-only registry and then trying to retrofit plugins later.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -402,6 +402,20 @@
 ### 2.8 Adapter Registry
+
+**External adapters (plugins):**
+Gas City supports loading runtime adapters implemented as external processes speaking JSON-RPC over stdin/stdout.
+Plugins are discovered from:
+- `workspace/.gc/plugins/`
+- `~/.gc/plugins/`
+
+Each plugin provides: `name`, `version`, `capabilities`, and implements the RuntimeAdapter RPC surface.
+If a plugin crashes, the controller restarts it with exponential backoff and marks dependent agents unhealthy.

@@ -2479,8 +2493,12 @@
 ### Phase 8: Additional Adapters + Conformance CLI (ongoing)
 
-**Goal:** Codex, Gemini, Agent SDK, Docker adapters + `gc test-adapter`.
+**Goal:** Codex, Gemini, Agent SDK, Docker adapters + plugin protocol + `gc test-adapter`.
 
 **Deliverables:**
 - Per-adapter package in `internal/runtime/`
+- JSON-RPC plugin host (`internal/runtime/pluginhost/`) and discovery from `.gc/plugins/`
 - Contract tests passing for each (happy + error paths)
 - `gc test-adapter <name>` CLI command for ecosystem certification
```

---

## 8) Replace `clear_context` Boolean With a Context Policy (Reliability + Long-Running Loops)

**Problem**  
`clear_context` is binary (`PLAN.md:521-525`): either you throw away useful continuity, or you risk context bloat and degraded performance in long-running loops. For autonomous agents (Ralph, worker pools), this becomes a reliability problem over time, not just a UX preference.

**Best fix**  
Introduce `context_policy` with explicit options, and make it cooperate with `completion_mode=file`:
- `clear`: current behavior (restart between tasks)
- `keep`: current behavior
- `summarize`: agent writes a rolling summary file; controller restarts and seeds with summary
- `window`: keep last N tasks’ summaries/artifacts

**Why it’s better**  
This gives you predictable long-running behavior and reduces “agent goes weird after 20 tasks.” It also makes performance/cost more stable by controlling context growth.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -521,9 +521,12 @@
 [agents.loop]                   # Task loop configuration
 enabled = false
 auto_execute = false            # GUPP: auto-start when work appears
-clear_context = false           # Clear agent context between tasks
+context_policy = "keep"         # "keep", "clear", "summarize", "window"
+window_tasks = 10               # Used by "window"
+summary_file = ".gc/summary.md" # Used by "summarize"
 poll_interval = "10s"           # How often to check for new tasks

@@ -688,7 +691,7 @@
-      if clear_context:
+      if context_policy == "clear":
         runtime.Stop(self.handle, graceful=true)
         self.handle = runtime.Start(ctx, self.config)
```

---

## Smaller but Important Spec Nits (Worth Fixing While Editing Anyway)

These are not new features, but they will cause confusion for implementers:
- `HookExecutor.Run()` assigns `events := h.bus.Subscribe(h)` even though `Subscribe` returns nothing (`PLAN.md:1275` vs `PLAN.md:2154-2157`).
- The “default priorities based on role” text conflicts with the “roles not hardcoded” claim (`PLAN.md:1049-1062` vs `PLAN.md:950`), even if you keep role-based defaults.

If you want, I can produce diffs for these as well, but the eight changes above are the highest leverage.
tokens used
122,990
**Vision Check (`../specs/gas-city-vision.md`)**  
The plan broadly satisfies the vision bullets: toolkit/not-one-shape (`PLAN.md:13-29`, `PLAN.md:643-1009`), progressive levels (`PLAN.md:581-626`), user-defined roles via external prompts (`PLAN.md:923-1009`), sandboxes/isolation (`PLAN.md:501-506`), uniform “factory worker” abstraction (`PLAN.md:56-88`). The remaining *vision gap* is “plugins” as a first-class extensibility surface: today it is “adapter registry + open question” (`PLAN.md:28`, `PLAN.md:402-432`, `PLAN.md:2557-2564`), not an actually usable plugin model.

Below are the biggest Round 2 improvements still missing, ordered by severity. Each includes rationale and a `git diff` style patch against the current `PLAN.md`.

---

## 1) Fix Task Claiming vs Assignment (Currently Inconsistent, Breaks Ralph + Pools)

**Problem**  
The Ralph loop lists “ready tasks assigned to me” (`PLAN.md:675`), but `Claim()` only succeeds when `Assignee == ""` and then sets `Assignee = agent` (`PLAN.md:1484`). That makes the loop logically impossible: tasks “assigned to me” are non-claimable by definition. Same issue for pools: tasks pre-assigned to a pool name (which is implied by `CountPending(assignee)` usage) can never be claimed under the current semantics. There is also a type mismatch: `MarkCompleted` expects a `TaskResult` but the loop passes `result.Output` (`PLAN.md:687` vs `PLAN.md:1432`).

**Best fix**  
Split “routing target” from “exclusive claim lock.” Keep `Assignee` as the *logical target* (agent name or pool name). Add `ClaimedBy` as the *actual* claimant identity (including pool instance). Change `Claim()` to set `ClaimedBy` and `Status`, not overwrite routing. Add a `TaskFilter.ClaimableBy` so loops can ask “what can I claim?” without encoding pool-matching rules in every caller.

**Why it’s better**  
This makes Ralph, Agent Teams, and “coordinator assigns work to a pool” all coherent. It also preserves important information: “this task was routed to pool devs” and “this specific worker instance claimed it,” which matters for debugging, retries, and for implementing fair scheduling.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -672,21 +672,21 @@
 LOOP:
-  tasks = taskBackend.ListReady(filter: {assignee: self.name, sortBy: "priority", sortOrder: "desc"})
+  tasks = taskBackend.List(filter: {status: Ready, claimable_by: self.identity, sortBy: "priority", sortOrder: "desc"})
   if len(tasks) == 0:
     sleep(poll_interval)
     goto LOOP
   task = tasks[0]  // highest priority
-  ok = taskBackend.Claim(task.id, self.name)  // Atomic claim
+  ok = taskBackend.Claim(task.id, self.identity)  // Atomic claim (records ClaimedBy)
   if !ok:
     goto LOOP  // Another agent claimed it first
   runtime.Assign(self.handle, task)
   WAIT_COMPLETION:
     result = runtime.WaitForResult(self.handle)  // Blocks with internal timeout
     if result.Status == Completed:
-      taskBackend.MarkCompleted(task.id, self.name, result.Output)
+      taskBackend.MarkCompleted(task.id, self.identity, result)
       if clear_context:
         runtime.Stop(self.handle, graceful=true)
         self.handle = runtime.Start(ctx, self.config)
       goto LOOP
     if result.Status == Failed:
       taskBackend.MarkFailed(task.id, self.name, result.Error)
       runtime.Restart(self.handle)
       goto LOOP

@@ -1421,34 +1421,43 @@
 type TaskBackend interface {
     // CRUD
     List(filter TaskFilter) ([]Task, error)
     Get(id string) (Task, error)
     Create(task Task) (string, error)  // Returns ID
     Update(id string, updates TaskUpdates) error
     CountPending(assignee string) (int, error)

     // Atomic operations for concurrency safety
-    Claim(id string, agent string) (bool, error)                     // Atomic: claim or return false
+    Claim(id string, claimant string) (bool, error)                  // Atomic: claim or return false
     MarkInProgress(id string, agent string) error
     MarkCompleted(id string, agent string, result TaskResult) error
     MarkFailed(id string, agent string, reason string) error
 }

 type Task struct {
@@
-    Assignee     string      // Agent name
+    Assignee     string      // Logical target: agent name or pool name (routing). Empty means unassigned/pullable.
+    ClaimedBy    string      // Claimant identity (e.g. "ws/proj/devs[2]"). Empty means unclaimed.
@@
 }

 type TaskFilter struct {
     Status    TaskStatus
     Assignee  string
+    ClaimableBy string // If set, backend returns only tasks this claimant is eligible to Claim()
     Project   string
     Labels    []string
     NoDeps    bool    // Only tasks with no unresolved dependencies
     SortBy    string  // "priority", "created", "updated"
     SortOrder string  // "asc", "desc"
 }

@@ -1484,7 +1493,20 @@
-**Claim semantics:** `Claim(id, agent)` atomically checks `Status == Ready && Assignee == ""`, sets `Status = InProgress, Assignee = agent`, and increments `Version`. If the task is already claimed, it returns `(false, nil)`. This prevents two agents from claiming the same task.
+**Claim semantics:** `Claim(id, claimant)` atomically checks:
+- `Status == Ready`
+- `ClaimedBy == ""`
+- AND the claimant is eligible:
+  - If `Assignee == ""`, anyone can claim (pull model).
+  - If `Assignee != ""`, the claimant must match the assignee (push model), including pool instances:
+    - Example: `Assignee = "devs"` is claimable by `"ws/proj/devs[0]"`, `"ws/proj/devs[1]"`, etc.
+
+On success it sets `Status = InProgress`, sets `ClaimedBy = claimant`, increments `Version`,
+and leaves `Assignee` unchanged (preserves routing). If already claimed, returns `(false, nil)`.

@@ -1508,7 +1530,7 @@
-    if task.Status != StatusReady || task.Assignee != "" {
+    if task.Status != StatusReady || task.ClaimedBy != "" {
         return false, nil  // Already claimed
     }
-    task.Status = StatusInProgress
-    task.Assignee = agent
+    task.Status = StatusInProgress
+    task.ClaimedBy = agent
     task.Version++
     task.UpdatedAt = time.Now()
```

---

## 2) Make Crash Recovery Realistic: “Adopt If Possible, Otherwise Restart”

**Problem**  
The plan promises “controller rediscovers running agents … and resumes control without restarting them” (`PLAN.md:1043`). That is not generally possible for subprocess-based adapters: after the controller dies, you cannot recover stdin/stdout pipes. Even for tmux you need an explicit “adopt/attach to existing runtime instance” operation that is not in the adapter contract (`PLAN.md:56-88`).

**Best fix**  
Introduce an *optional* adoption interface. The controller uses it when available (tmux, docker). If unsupported, the controller explicitly restarts the agent (and records that in the registry/events).

**Why it’s better**  
This makes the promise true for adapters that can support it, and avoids “spec says X but implementation cannot” for ones that cannot. It also clarifies which state you need to persist in `AgentHandle.Metadata` to make adoption work.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -90,6 +90,16 @@
 type ReadinessProbe interface {
     IsReady(handle AgentHandle) (bool, error)
     WaitReady(ctx context.Context, handle AgentHandle) error
 }
+
+// Adoption is an optional interface adapters can implement to support
+// controller crash recovery without forcing an agent restart.
+type Adoption interface {
+    // Adopt attaches to an already-running runtime instance described by the
+    // persisted handle (PID/session/container ID in handle.Metadata).
+    // Returns an updated handle or ErrNotSupported.
+    Adopt(ctx context.Context, persisted AgentHandle) (AgentHandle, error)
+}

@@ -1043,7 +1053,13 @@
-**Crash recovery:** If the controller crashes, agents continue running (they are independent OS processes). On next `gc start`, the controller rediscovers running agents via the persisted agent registry (`.gc/agents/*.json`) and resumes control without restarting them.
+**Crash recovery:** If the controller crashes, agents may continue running (they are independent OS processes). On next `gc start`, the controller loads persisted handles from `.gc/agents/*.json` and:
+1. If the adapter implements `Adoption`, it calls `Adopt()` to re-attach without restart (tmux, docker).
+2. Otherwise it restarts the agent with `Start()` and publishes an event noting the restart (subprocess-style runtimes cannot recover I/O pipes).
```

---

## 3) Make Spec Internals Consistent: `completion_mode` + AgentConfig Fields Must Match Usage

**Problem**  
The spec defines `completion_mode` in `[agents.runtime_config]` (`PLAN.md:102-117`), but it is missing from the “Full Schema” (`PLAN.md:509-514`). Separately, `AgentConfig` is missing fields that the plan later treats as present (notably health timeouts in contract tests and pool readiness usage).

**Best fix**  
Add the common `completion_mode` knobs to the schema, and update `AgentConfig` in Section 2.3 to include the fields the rest of the spec assumes exist (health/loop/pool/prompts/hooks and completion config). Keep `RuntimeConfig map[string]any` for adapter-specific extras.

**Why it’s better**  
This eliminates “paper cuts” where implementers build the wrong structs or omit key fields, and it makes the contract tests and controller algorithms materially implementable.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -177,6 +177,30 @@
 type AgentConfig struct {
     Name          string
     Role          string
     Runtime       string
     Project       string            // Empty for workspace scope
+    Scope         string            // "workspace" or "project"
+    Ephemeral     bool
+    Isolation     string            // "none"|"worktree"|"directory"|"container"
+
     Command       string            // Primary command (e.g., "claude")
     Args          []string
     Env           map[string]string
     WorkDir       string
     SystemPrompt  string            // Path or inline prompt
     NudgePrompt   string            // Initial message to send
+
+    // Completion detection (Section 2.2.1)
+    CompletionMode     string // "file"|"prompt"|"exit"
+    PromptRegex        string // used by prompt mode (optional)
+    ResultFileTemplate string // used by file mode (optional)
+
+    // Orchestration policies (schema sections)
+    Health   AgentHealthConfig
+    Pool     AgentPoolConfig
+    Loop     AgentLoopConfig
+    Prompts  AgentPromptConfig
+    Hooks    AgentHookConfig
+
     RuntimeConfig map[string]any    // Adapter-specific settings
     DependsOn     []string          // Agent names that must start first
 }

@@ -509,6 +533,12 @@
 [agents.runtime_config]         # Adapter-specific settings
 command = "claude"              # Command to run
 args = ["--dangerously-skip-permissions"]
 env = { KEY = "${MY_SECRET}" }  # Extra env vars (supports ${VAR} expansion)
+completion_mode = "prompt"      # "file", "prompt", or "exit" (Section 2.2.1)
+prompt_regex = ""              # Optional: prompt detection pattern for "prompt" mode
+result_file_template = ""       # Optional: e.g. ".gc/result-{{.TaskID}}.json" for "file" mode
 resume_flag = "--resume"        # How this runtime resumes sessions
 resume_style = "flag"           # "flag" or "subcommand"
```

---

## 4) Remove Hidden Role Coupling: Make Monitoring + Startup Ordering Explicit (Toolkit, Not “Blessed Roles”)

**Problem**  
The plan claims “no hardcoded role logic” (`PLAN.md:950`), but it still hardcodes role-based priorities (`PLAN.md:1051-1062`), makes Level 6 implicit on “supervisor role exists” (`PLAN.md:571-573`, `PLAN.md:613-615`), and enforces role-count limits (`PLAN.md:636-637`). That is fine for a product, but it is at odds with the “build any town shape” toolkit posture.

**Best fix**  
Add explicit knobs:
- `startup_priority` (per agent) so startup ordering is not coupled to role strings.
- `[monitoring]` (Level 6) so monitoring is a configured capability, not inferred from a magic role.
- Relax “only one coordinator/supervisor” from hard errors to either warnings or an optional “strict mode.”

**Why it’s better**  
This preserves progressive levels while removing “reserved role” behavior from the core. It makes the SDK more general and less surprising for Level 8 users building non-Gas-Town topologies.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -498,6 +498,8 @@
 role = "worker"                 # "coordinator", "supervisor", "observer", "integrator",
                                 # "worker", "agent", "service", or any custom string
+startup_priority = 2            # Optional. Lower starts earlier when depends_on is empty.
 scope = "project"               # "workspace" (one total) or "project" (one per project)
 ephemeral = false               # Whether this agent is created/destroyed per-task

@@ -571,9 +573,11 @@
-# === LEVEL 6: Monitoring (implicit when supervisor role exists) ===
-# No extra config section needed — monitoring is enabled when a
-# supervisor agent is configured.
+# === LEVEL 6: Monitoring ===
+[monitoring]
+enabled = false
+patrol_interval = "30s"

@@ -613,8 +617,8 @@
-    if level >= 5 && hasSupervisor(cfg) {
+    if level >= 5 && cfg.Monitoring != nil && cfg.Monitoring.Enabled {
         level = 6 // Level 6: Monitoring
     }

@@ -636,8 +640,8 @@
-| Coordinator limit | At most 1 coordinator per workspace | "multiple coordinators not supported" |
-| Supervisor limit | At most 1 supervisor per workspace | "multiple supervisors not supported" |
+| Coordinator limit | (warning) Multiple coordinators allowed; `gc init` uses 1 by default | "warning: multiple coordinators configured" |
+| Supervisor limit | (warning) Multiple monitors allowed; `gc init` uses 1 by default | "warning: multiple supervisors configured" |

@@ -1049,19 +1053,13 @@
-Agents start according to a dependency DAG. If no `depends_on` is specified, agents use default priority groups based on role.
+Agents start according to a dependency DAG. If no `depends_on` is specified, agents use `startup_priority` (default 2).

-// Default priorities (used when depends_on is not specified)
-var defaultPriority = map[string]int{
-    "supervisor":  0,
-    "coordinator": 1,
-    ...
-}
-// Custom roles default to priority 2 (same as per-project agents)
+// Startup priority is configured per agent. `gc init` chooses sensible defaults for known shapes.
```

---

## 5) Either Specify Batches (Convoys) or Remove Them From Schema (Dangling Feature Right Now)

**Problem**  
`[batches]` is present in the schema and the Gas Town example (`PLAN.md:575-578`, `PLAN.md:863-866`), but there is no batch data model, no persistence, no events, and no CLI. As written, it is an unimplemented promise and undermines “spec is implementable.”

**Best fix**  
Define batches as a controller-side, backend-agnostic grouping stored in `.gc/batches/` (so it works with GitHub issues, beads, filesystem tasks). Make `auto_track` create a batch whenever a workflow dispatches multiple tasks.

**Why it’s better**  
This gives users a compelling “orchestration view” without forcing every TaskBackend to grow a batch concept. It also helps with multi-project operations (the whole point of Level 7).

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -1540,6 +1540,62 @@
 ---
 
 ## 9. Messaging System
+
+@@
+## 8. Task System
+...
+### 8.3 Batch (Convoy) Tracking
+
+Batches are a workspace-local grouping of tasks for tracking multi-task dispatches (Gas Town "convoys").
+Batches are stored in `.gc/batches/<batch-id>.json` and reference task IDs in the configured TaskBackend.
+This keeps batch tracking backend-agnostic (works with beads, GitHub Issues, filesystem).
+
+```go
+type BatchStatus string
+const (
+    BatchRunning   BatchStatus = "running"
+    BatchCompleted BatchStatus = "completed"
+    BatchFailed    BatchStatus = "failed"
+)
+
+type Batch struct {
+    ID          string
+    Title       string
+    Description string
+    Project     string   // optional
+    TaskIDs     []string
+    Status      BatchStatus
+    CreatedAt   time.Time
+    UpdatedAt   time.Time
+}
+```
+
+**auto_track behavior (`[batches] auto_track = true`):**
+1. When a workflow dispatches 2+ tasks, create a new Batch referencing those task IDs.
+2. Batch status is derived from task statuses:
+   - Completed if all tasks Completed
+   - Failed if any task Failed (configurable later)
+   - Running otherwise
+
+**CLI additions:**
+`gc batch list`, `gc batch show <id>`, `gc batch watch <id>`
```

(If you do *not* want to ship batches in v1, the better alternative is to remove `[batches]` from the schema and Gas Town example until specified.)

---

## 6) Make the Event Bus Safe Under Load + Define How “Agent Logs” Become Events

**Problem**  
The bus still spawns one goroutine per optional subscriber per event (`PLAN.md:1267-1270`). Under bursty output (agent logs, workflow step spam), this becomes unbounded goroutines and can destabilize the controller. Separately, the plan claims “structured logger subscribes to all events + agent logs” (`PLAN.md:1309`) but there is no event type or ingestion path for agent output.

**Best fix**  
1. Replace “goroutine per event” with per-subscriber bounded queues and a single worker goroutine per subscriber (preserves ordering, bounds memory/cpu).  
2. Introduce `agent.output` events and a controller “log pump” that reads from `RuntimeAdapter.CaptureOutput()` (or polls `ReadOutput()` for snapshot adapters) and publishes normalized log line events.

**Why it’s better**  
This is the difference between “works in a demo” and “can run overnight with 10 agents.” It also makes `gc agent logs --follow` and `gc activity --follow` implementable without adapter-specific hacks.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -1210,6 +1210,7 @@
 const (
@@
     EventAgentReady     EventType = "agent.ready"
+    EventAgentOutput    EventType = "agent.output" // normalized log lines/frames
@@
 )

@@ -1238,12 +1239,19 @@
 type EventBus struct {
     mu          sync.RWMutex
     sequence    int64
-    critical    []EventSubscriber  // Block on these (supervisor, structured logger)
-    optional    []EventSubscriber  // Fire-and-forget (CLI feed, metrics)
+    critical    []*subscriberState // bounded queue + single worker per subscriber
+    optional    []*subscriberState // bounded queue + single worker per subscriber
     buffer      *ring.Buffer[Event]  // Ring buffer (10k events) for catch-up on subscribe
 }
+
+type subscriberState struct {
+    sub   EventSubscriber
+    ch    chan Event
+    drop  bool // true for optional subscribers: drop when full
+}

@@ -1267,10 +1275,11 @@
-    // Optional subscribers: fire-and-forget
-    for _, sub := range optionals {
-        go sub.OnEvent(event)
-    }
+    // Optional subscribers: enqueue (drop when full), no unbounded goroutines
+    for _, st := range optionals {
+        select { case st.ch <- event: default: /* drop */ }
+    }

@@ -1308,7 +1318,8 @@
-| Structured logger | Critical | All events + agent logs | `workspace/.gc/logs.jsonl` (attributed with agent_id) |
+| Structured logger | Critical | All events + agent.output | `workspace/.gc/logs.jsonl` (attributed with agent_id) |
 | Hook executor | Critical | Configurable per event type | Runs user-defined shell commands |
```

---

## 7) Turn “Plugins” From an Open Question Into a Usable Extensibility Surface

**Problem**  
Right now, adding a runtime adapter requires modifying Go and recompiling (`PLAN.md:402-432`). The vision explicitly calls out plugins (`../specs/gas-city-vision.md:24-25`). The plan mentions JSON-RPC only as a future “Phase 8” open question (`PLAN.md:2557`, `PLAN.md:2479-2487`), which delays the most “toolkit-y” part of the toolkit.

**Best fix**  
Define the plugin boundary now (protocol + discovery), even if only a minimal loader ships early:
- Plugin discovery: workspace-local `.gc/plugins/` plus user-global `~/.gc/plugins/`.
- Protocol: JSON-RPC over stdio with version handshake and capability advertisement.
- Support at least runtime adapters; ideally also task/message backends later.

**Why it’s better**  
This is what lets Level 8 users wire in internal models, bespoke sandboxes, or custom execution substrates without forking Gas City. It also reduces long-term architectural risk: you avoid baking in an internal-only registry and then trying to retrofit plugins later.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -402,6 +402,20 @@
 ### 2.8 Adapter Registry
+
+**External adapters (plugins):**
+Gas City supports loading runtime adapters implemented as external processes speaking JSON-RPC over stdin/stdout.
+Plugins are discovered from:
+- `workspace/.gc/plugins/`
+- `~/.gc/plugins/`
+
+Each plugin provides: `name`, `version`, `capabilities`, and implements the RuntimeAdapter RPC surface.
+If a plugin crashes, the controller restarts it with exponential backoff and marks dependent agents unhealthy.

@@ -2479,8 +2493,12 @@
 ### Phase 8: Additional Adapters + Conformance CLI (ongoing)
 
-**Goal:** Codex, Gemini, Agent SDK, Docker adapters + `gc test-adapter`.
+**Goal:** Codex, Gemini, Agent SDK, Docker adapters + plugin protocol + `gc test-adapter`.
 
 **Deliverables:**
 - Per-adapter package in `internal/runtime/`
+- JSON-RPC plugin host (`internal/runtime/pluginhost/`) and discovery from `.gc/plugins/`
 - Contract tests passing for each (happy + error paths)
 - `gc test-adapter <name>` CLI command for ecosystem certification
```

---

## 8) Replace `clear_context` Boolean With a Context Policy (Reliability + Long-Running Loops)

**Problem**  
`clear_context` is binary (`PLAN.md:521-525`): either you throw away useful continuity, or you risk context bloat and degraded performance in long-running loops. For autonomous agents (Ralph, worker pools), this becomes a reliability problem over time, not just a UX preference.

**Best fix**  
Introduce `context_policy` with explicit options, and make it cooperate with `completion_mode=file`:
- `clear`: current behavior (restart between tasks)
- `keep`: current behavior
- `summarize`: agent writes a rolling summary file; controller restarts and seeds with summary
- `window`: keep last N tasks’ summaries/artifacts

**Why it’s better**  
This gives you predictable long-running behavior and reduces “agent goes weird after 20 tasks.” It also makes performance/cost more stable by controlling context growth.

```diff
diff --git a/PLAN.md b/PLAN.md
--- a/PLAN.md
+++ b/PLAN.md
@@ -521,9 +521,12 @@
 [agents.loop]                   # Task loop configuration
 enabled = false
 auto_execute = false            # GUPP: auto-start when work appears
-clear_context = false           # Clear agent context between tasks
+context_policy = "keep"         # "keep", "clear", "summarize", "window"
+window_tasks = 10               # Used by "window"
+summary_file = ".gc/summary.md" # Used by "summarize"
 poll_interval = "10s"           # How often to check for new tasks

@@ -688,7 +691,7 @@
-      if clear_context:
+      if context_policy == "clear":
         runtime.Stop(self.handle, graceful=true)
         self.handle = runtime.Start(ctx, self.config)
```

---

## Smaller but Important Spec Nits (Worth Fixing While Editing Anyway)

These are not new features, but they will cause confusion for implementers:
- `HookExecutor.Run()` assigns `events := h.bus.Subscribe(h)` even though `Subscribe` returns nothing (`PLAN.md:1275` vs `PLAN.md:2154-2157`).
- The “default priorities based on role” text conflicts with the “roles not hardcoded” claim (`PLAN.md:1049-1062` vs `PLAN.md:950`), even if you keep role-based defaults.

If you want, I can produce diffs for these as well, but the eight changes above are the highest leverage.
