# Tutorial 02 — Looping with Ralph

> **Status:** NOTES — not yet a tutorial

---

## Problem

From Tutorial 01, the mayor works one bead and you have to hand it the next.
If you have a backlog of ten beads, that's ten nudges. You want to fill a
backlog and walk away. The mayor should drain it — one bead at a time, each
with a clean context.

---

## What This Tutorial Adds

- `[agents.loop]` config — turns the mayor into a polling loop
- The **hook** system — atomic bead claiming so an agent owns its work
- **Context survival** — when context fills up mid-task, the hooked bead
  persists and the fresh session picks it up

---

## Hook System (moved from Tutorial 01)

When the mayor (or any agent) picks up a bead in a loop, it **hooks** the
bead — atomically claiming it. This matters as soon as you have more than
one thing going on:

- `gc hook <bead-id>` — attach a bead to the agent's hook
- Status goes from `open` → `hooked`
- **Hooking is atomic** — if two agents tried to claim the same bead, exactly
  one would succeed. The other gets a conflict error and moves on. This is
  how Gas City prevents double-work when you scale to many agents.
- The bead's assignee field records who hooked it
- When work is done: `gc bead close` → status goes `hooked` → `closed`

Full lifecycle: `open` → `hooked` → `closed`

In Tutorial 01 the lifecycle was simpler (open → active → closed) because
there was only one agent. Hooks become essential when the loop introduces
repeated claim/release cycles, and critical when Tutorial 03 adds multiple
agents competing for work.

---

## Context Survival (moved from Tutorial 01)

This is where beads really prove their worth. When the loop is running and
the mayor's context fills up mid-task:

1. The mayor's session ends (context exhausted)
2. The bead is still `hooked` to the mayor — the assignment persists
3. Any code changes the mayor made are still in the working directory
4. The city restarts the mayor with a fresh session
5. The mayor's startup checks for hooked work, finds the bead
6. It examines the codebase to see where the last session left off
7. Continues from there

**The aha moment:** The mayor ran out of context, got a fresh session, and
didn't miss a beat — because the bead knew what was assigned and the code on
disk showed what was done. Context windows come and go; beads persist.

In a loop, this happens naturally. Each bead gets a clean context. If a bead
is too big for one context window, the hook system ensures the fresh session
picks it back up instead of grabbing the next thing in the queue.

---

## Simulated Interaction Sketch

```shell
# Fill the backlog
$ gc bead create "Implement 3-disk solver"
Created bead: gc-2  (status: open)

$ gc bead create "Add animation to disc moves"
Created bead: gc-3  (status: open)

$ gc bead create "Write unit tests for the solver"
Created bead: gc-4  (status: open)

$ gc bead ready
ID    STATUS   TITLE
gc-2  open     Implement 3-disk solver
gc-3  open     Add animation to disc moves
gc-4  open     Write unit tests for the solver
```

Tell the mayor to start looping (or it starts automatically with config):

```
Mayor claims gc-2, works it, closes it.
Mayor claims gc-3, works it, closes it.
Mayor claims gc-4, works it, closes it.

$ gc bead list
ID    STATUS   ASSIGNEE   TITLE
gc-1  closed   mayor      Build a Tower of Hanoi app
gc-2  closed   mayor      Implement 3-disk solver
gc-3  closed   mayor      Add animation to disc moves
gc-4  closed   mayor      Write unit tests for the solver
```

You filled the backlog. The mayor drained it. Each bead got a clean context.

---

## Configuration

What needs to be added to the city config to enable looping. Exact format TBD
based on whether we keep TOML or move to the new config model from Tutorial 01.

From the spec, the loop config looks like:

```toml
[agents.loop]
enabled = true
auto_execute = true
poll_interval = "30s"
```

But with the new DX (city-as-directory, `gc rig add`, default mayor), this
might become a setting you toggle:

```shell
$ gc config set loop.enabled true
```

Or it might be part of the built-in config tiers:

```shell
$ gc init ralph
# Sets up a city with loop enabled
```

---

## Fodder from Tutorial 01

The `gc init` wizard and hello-world.toml config — these may become part of
how we explain the different built-in configurations:

```
$ gc init
Welcome to Gas City SDK!

Example configs available:
  > hello-world.toml -- Single agent, single task (simplest)
    ralph.toml       -- Single agent with task loop
    ccat.toml        -- Coordinator + worker pool (Agent Teams)
    gastown.toml     -- Full multi-project orchestration (Gas Town)
    (custom)         -- Start from scratch
```

The TOML-based hello-world config for reference:

```toml
# hello-world.toml
[workspace]
name = "my-project"

[projects.main]
repo = "."

[tasks]
backend = "beads"

[[agents]]
name = "worker"
```

And the TOML config breakdown:

- **`[workspace]`** names your workspace. This becomes the prefix for bead IDs.
- **`[projects.main]`** points at your Git repo. `"."` means the current
  directory. You can add more projects later (Tutorial 04d does exactly that).
- **`[tasks]`** activates the beads task store. Without this section, Gas City
  has no way to track work. With it, every task gets a persistent bead that
  outlives any agent session.
- **`[[agents]]`** defines one agent named "worker." The provider (Claude Code,
  Codex, Gemini, etc.) is auto-detected from whatever's installed on your PATH.

> **What's NOT here matters too.** No `[agents.loop]` — so the agent won't
> poll for work automatically. No `[messaging]` — there's nobody to message.
> No `[daemon]` — no health monitoring. Gas City activates subsystems based
> on what's in your config. Right now, that's just an agent and a task store.

---

## Open Questions

- Does "ralph" become a named config tier, or is the tutorial just about
  adding loop config to what you already have from Tutorial 01?
- How does `gc init ralph` relate to the new city-as-directory model?
- Is the loop config a TOML section, a `gc config set` command, or something
  else in the new DX?
