Loaded cached credentials.
## Gas City SDK - Post-Hardening Review (Round 4)

**Reviewer:** Gemini
**Date:** 2026-02-14
**Verdict:** Implementation readiness: **YES**

---

### 1. Overall Assessment

This is an exceptionally well-architected and comprehensive technical specification. The progression from a monolithic application (Gas Town) to a composable SDK (Gas City) is clearly articulated and technically sound. The document has significantly benefited from previous review rounds and the recent hardening phase, which introduced formal guarantees that are both correct and well-integrated.

The design exhibits a high degree of maturity, proactively addressing complex issues like concurrency, crash recovery, and resource management. The core abstractions — particularly the `RuntimeAdapter` and the tiered `EventBus` — are robust and provide the right seams for extensibility. The phased implementation plan is realistic and de-risks the project.

The specification is ready for implementation. The findings below are minor and intended for clarification rather than as blockers.

---

### 2. Review of Formal Guarantees

The formal properties and correctness guarantees are a standout feature of this specification. My review confirms the following:

-   **Correctness:** All stated guarantees (e.g., Task Claim Atomicity, Startup Rollback Safety, Pool Bounds, Single Controller) are formulated correctly and represent critical properties for a robust orchestration system.
-   **Integration:** The guarantees are not merely asserted; they are woven directly into the design of the relevant components. The specification provides clear implementation strategies (e.g., using `flock` for filesystem locks, topological sorting for startup) that directly support and enforce these properties. The accompanying proofs are logical and sound.
-   **Value:** These properties provide a strong contract for implementers and a high degree of confidence in the system's stability and correctness.

The hardening phase was successful. The formal guarantees are a cornerstone of this specification's quality.

---

### 3. Findings

#### CRITICAL
None.

#### HIGH
-   **Finding 1: Ambiguity in `prompt` Completion Mode Configuration**
    -   **Location:** Section 2.2.1 (Agent Task Protocol), Section 3.2 (Full Schema)
    -   **Issue:** The `prompt` completion mode is a necessary heuristic for interactive agents, but the specification does not define *how* the prompt pattern itself is configured. For a generic adapter like `subprocess`, this is a critical missing piece of configuration.
    -   **Recommendation:** Add a `completion_prompt_pattern` (string) field to the `[agents.runtime_config]` table in the `gas-city.toml` schema. The documentation for the `subprocess` and `claude-code` adapters should clarify that they use this pattern to detect task completion when `completion_mode` is set to `"prompt"`.

#### MEDIUM
-   **Finding 2: Minor Inefficiency in PoolManager Scale-Up Logic**
    -   **Location:** Section 4.2 (Pool scaling algorithm)
    -   **Issue:** The algorithm releases its lock after deciding to scale up but before starting the new agent. This creates a small window where the number of pending tasks could decrease (due to another worker becoming free), potentially making the new agent superfluous upon start.
    -   **Analysis:** This is not a correctness bug, as the idle agent would eventually be scaled down. It is a minor resource efficiency issue. The design correctly prioritizes not holding a lock during a potentially slow I/O operation (`Start`).
    -   **Recommendation:** This behavior is an acceptable trade-off. Add a sentence to the "Pool scaling algorithm" section acknowledging this possibility and stating it's an intentional design choice to avoid long-held locks. No code change is required.

#### LOW
-   **Finding 3: Missing Link in `Adopter` Interface Metadata**
    -   **Location:** Section 2.2 (The RuntimeAdapter Interface), Section 5.0 (Workspace Controller)
    -   **Issue:** The `Adopter.Adopt` method takes a `metadata` map, but the source of this metadata isn't explicitly stated. The spec mentions crash recovery uses persisted files in `.gc/agents/`, but the connection is implicit.
    -   **Recommendation:** In Section 5.0 (Crash Recovery), add a sentence clarifying that the `metadata` map passed to the `Adopt` method is populated from the `AgentHandle.Metadata` field, which is persisted in the agent's state file (`.gc/agents/<addr>.json`) by the `AgentRegistry`.

-   **Finding 4: Inflexible Event Bus Subscription Replay**
    -   **Location:** Section 6.2 (Bus Implementation)
    -   **Issue:** The `EventBus.Subscribe` method unconditionally replays the entire history buffer to every new subscriber. Consumers that only care about future events (e.g., a real-time UI feed opened mid-session) receive this potentially large replay unnecessarily.
    -   **Recommendation:** For v1, this is acceptable. For a future version, consider modifying the signature to `Subscribe(sub EventSubscriber, options SubscribeOptions)` where `options` could include a `ReplayHistory bool` flag (defaulting to true) to allow consumers to opt out of the replay.

---

### 4. Convergence Verdict: YES

The specification is robust, detailed, and implementation-ready. It successfully translates the high-level vision for Gas City into a concrete and technically sound plan. The identified issues are minor and can be addressed as points of clarification during development. The project should proceed to the implementation phase.
