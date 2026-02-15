## Gas City SDK - Post-Hardening Review (Round 4)

**Reviewer:** Claude Opus 4.6
**Date:** 2026-02-14
**Spec Version:** 0.5.0
**Verdict:** Implementation readiness: **YES**

---

### 1. Formal Content Quality

The spec contains 12 formal guarantees and invariants spread across Sections 2, 3, 5, 6, 8, and 18.

#### P1-P6 (Adapter Correctness Properties, Section 2.5)
- P1 (Idempotent Stop): Correct. Contract test validates.
- P2 (Liveness After Start): Sound existential guarantee.
- P3 (Graceful Degradation): Correct. Contract test validates.
- P4 (Bounded Resource Cleanup): **MEDIUM** -- lacks corresponding contract test.
- P5 (Thread Safety): Correct. Contract test validates concurrent access.
- P6 (Start Uniqueness): Correct. Contract test validates.

#### Monotonicity Invariant (Section 3.3 + 18.1)
- Invariant holds but proof conflates "config section" with "config characteristic" -- Level 6 is detected via agent role presence, not a new section.
- **Rating: LOW**

#### Single Controller Guarantee (Section 5.0)
- Proof correct. `flock(LOCK_EX|LOCK_NB)` semantics sound for local-only execution scope.
- **Rating: No issue.**

#### Startup Ordering Invariant (Section 5.1)
- Invariant is a point-in-time assertion, not continuous. A dependency could crash between confirmation and dependent's start. Supervisor handles recovery.
- **Rating: LOW**

#### No Work Loss on Shutdown (Section 5.2)
- Race condition at drain-timeout boundary: concurrent `MarkCompleted` and `requeueInterruptedTasks` could conflict. Needs CAS semantics specification.
- **Rating: MEDIUM**

#### Critical Delivery and Sequence Monotonicity (Section 6.2)
- Both proofs correct. Replay mechanism in Subscribe handles any gap from lock release.
- **Rating: No issue.**

#### No Double-Execution (Section 8.1)
- Correct for filesystem (flock) and SQL backends. GitHub Issues backend "two-step atomic" verify mechanism is not fully specified.
- **Rating: MEDIUM**

#### Dependency Correctness (Section 8.1)
- Proof needs to clarify lock/transaction scope covers the *dependent* task's status transition, not just the completing task.
- **Rating: LOW**

#### Pool Bounds Invariant (Section 18.3)
- **HIGH** -- Formal claim of continuous mutex-protected enforcement is contradicted by two-phase (compute-then-execute) implementation that explicitly releases the lock between phases. Invariant should be restated as "eventually consistent within one scaling cycle."

#### Startup Rollback Safety (Section 18.6)
- Rollback function lacks graceful-to-force escalation, unlike shutdown sequencer.
- **Rating: MEDIUM**

---

### 2. Integration Quality

- Formal guarantees are well-integrated into their respective sections, not appended as afterthought.
- **HookConfig has three inconsistent representations** (struct, TOML schema, executor event dispatch). **HIGH.**
- AgentState comment lists "Running" but enum has "Starting". **LOW.**
- `TaskResult.Status` uses overly broad `TaskStatus` type. **MEDIUM.**
- `resume_flag` / `resume_style` config fields are unreferenced. **MEDIUM.**

---

### 3. Findings Summary

#### CRITICAL
None.

#### HIGH (2 issues)
- **H1:** Pool bounds invariant proof incorrect due to two-phase lock release.
- **H2:** HookConfig has three inconsistent representations (struct, TOML, executor).

#### MEDIUM (6 issues)
- M1: P4 (Bounded Resource Cleanup) lacks contract test.
- M2: No-work-loss shutdown guarantee has race at drain-timeout boundary.
- M3: No-double-execution proof incomplete for GitHub Issues backend.
- M4: Startup rollback lacks graceful-to-force escalation.
- M5: `TaskResult.Status` uses overly broad `TaskStatus` type.
- M6: `resume_flag` / `resume_style` config fields unreferenced.

#### LOW (4 issues)
- L1: Monotonicity proof conflates "config section" with "config characteristic" for Level 6.
- L2: Startup ordering invariant is point-in-time, not continuous.
- L3: Dependency correctness proof ambiguous on lock scope.
- L4: AgentState comment lists "Running" but enum has "Starting."

---

### 4. Convergence Verdict: YES

The spec is ready for implementation. No CRITICAL issues. The architecture is fundamentally sound. The two HIGH issues are reconcilable during implementation. Cross-model convergence is clear: Round 1 found architectural gaps, Round 2 found interface inconsistencies, Round 3 found naming issues, Round 4 found formal proof precision issues but no new architectural concerns. The trajectory is clearly convergent.
