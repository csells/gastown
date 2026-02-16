# Tutorial 01 — Hello, Gas City

Let's say that you're using Claude Code on a significant feature implementation.
You've described the feature, pointed the agent at the right files, and it's
making progress. Then — mid-flight — the context window fills up. The session is
either over or you're at the mercy of the compactation to save the important
details. This is the fundamental problem with AI coding agents: their memory is
the context window, and context windows are finite.

Gas City fixes this with **beads** — tracked work units that persist outside the
agent. When the agent uses beads to record what's done and what's left, running
out of context is no longer catastrophic. In fact, it can be beneficial to clear
out the context before it rots. A fresh session queries beads and picks up right
where the last one left off. The state is in the store, not in the agent's head.

This tutorial builds the simplest possible Gas City workspace: one agent, one
task, and the beads system that makes context survival work.

---

## Starting your city

First, start by installing Gas City. We do that on macOS with Home Brew:

```shell
$ brew install gascity
```

A city is a particular set of rules for how your orchestration works and a set
of projects configured for that orchestration. The configuration for a city is
stored in a folder on your computer set aside for that purpose. You can define
multiple cities with multiple configurations, but we'll start with just one for
now. By convention, each city goes into your home folder. You can start one up
with the default configuration like so:

```shell
$ mkdir ~/bright-lights
$ cd ~/bright-lights
$ gc start ~/bright-lights

Welcome to Gas City!
To configure your new city, add a `settings.yaml` file. To get started with one
of the built-in configurations, use `gc init`.
```

Starting a city uses the configuration to ensure that the agents are ready to do
their work and a whole host of other things. You can update the configuration at
any time, stop and restart your city for a new configurations to take affect.
If you specify no configuration, you'll get the default.

## Adding a project

To associate a project (called a "rig") with a city, you add it:

```shell
$ cd ~/bright-lights
$ gc rig add ~/projects/tower-of-hanoi
$ gc rig list

Rigs in /Users/csells/bright-lights:

  tower-of-hanoi:
    Agents: [mayor]
```

The "mayor" is the first agent of your city and is who you'll talk to do the
planning and work.

## Create a Task

Before starting the agent, give it something to do by creating a bead that
represents a task. You can do that manually or talk to the mayor:

```shell
# create the bead manually
$ gt bead create "build a Tower of Hanoi app"
Created bead: gc-1  (status: open)

# list the open beads (use this less)
$ gc bead list

# list the beads ready to work on (use this more)
$ gc bead ready

# talk to the mayor
gc agent attach mayor

[show CC gunk]

Mr. Mayor, can you create a bead to build a Tower of Hanoi app? thanks!

[show mayor response]

Can you list the ready beads?

[show mayor response]
```

You can learn and use the entire "gc" CLI if you like. Or you can talk to any of
the agents in Gas City, who know how to use that CLI for you. You can attach to
any agent by name. This city has only one agent in it by default, the mayor. The
act of "attaching" to the mayor brings up the single instance of the mayor,
either in a new tmux session or bringing the tmux session to the forefront if it
doesn't already exist.

Either way, the bead exists now, independent of any agent. A new bead starts
with a status of `open` — available for claiming. No assignee yet.

```
$ gc bead show mp-1
ID:       mp-1
Status:   open
Type:     task
Title:    Fix the off-by-one error in pagination.go line 42
Project:  main
Created:  2026-02-15 10:30:00
Assignee: —
```

> **Three things to notice:**
>
> 1. The bead has an ID (`mp-1`) derived from your workspace name. Every bead
>    in this workspace will be `mp-*`.
> 2. The bead has a status lifecycle: `open` → `hooked` → `closed`. Right now
>    it's `open`.
> 3. The bead is stored in `.beads/` inside your repo — not in the agent's
>    context window, not in memory, not anywhere ephemeral. It's on disk.

---

TODO: STARTHERE

## Start the Agent

```
$ gc start
Starting agent 'worker' (auto-detected: claude)...
Agent 'worker' is running.
```

Gas City detected Claude Code on your PATH and started it in a tmux session.
The agent receives a startup prompt that tells it to check for available work.
It finds `mp-1`, claims it, and starts working:

```
$ gc bead list
ID    STATUS   ASSIGNEE   TITLE
mp-1  hooked   worker     Fix the off-by-one error in pagination.go line 42
```

The status changed from `open` to `hooked`, and the assignee is now `worker`.
**Hooking is atomic** — if two agents tried to claim this bead simultaneously,
exactly one would succeed. The other would get a conflict error and move on.
This is how Gas City prevents double-work.

The agent is now working on the fix. You can attach to its tmux session to
watch (`gc agent attach worker`), or just let it run.

When the agent finishes, it closes the bead:

```
$ gc bead list
ID    STATUS   ASSIGNEE   TITLE
mp-1  closed   worker     Fix the off-by-one error in pagination.go line 42
```

Done. The bead records that the work happened, who did it, and when it closed.

> **What just happened?** Three Gas City primitives worked together:
>
> - **Config** told Gas City that one agent exists and beads are the task backend
> - **Agent Protocol** started the agent and delivered its startup prompt
> - **Task Store (Beads)** tracked the work unit through its lifecycle:
>   `open` → `hooked` → `closed`
>
> No messaging, no formulas, no health monitoring. Just the foundation.

---

## Survive a Context Reset

That first run was straightforward — but it doesn't demonstrate why beads
matter. For that, the agent needs to lose its context mid-task.

Create another bead, something that takes real work:

```
$ gc bead create "Refactor the auth middleware to support JWT and session tokens"
Created bead: mp-2  (status: open)

$ gc start
Starting agent 'worker' (auto-detected: claude)...
Agent 'worker' is running.
```

The agent claims `mp-2` and starts working. It's reading files, planning
the refactor, making changes. Then — partway through — the context window
fills up. The agent session ends.

```
$ gc bead list
ID    STATUS   ASSIGNEE   TITLE
mp-1  closed   worker     Fix the off-by-one error in pagination.go line 42
mp-2  hooked   worker     Refactor the auth middleware to support JWT and session tokens
```

The agent is gone, but the bead is still `hooked`. The work-in-progress is
recorded: the bead knows it was assigned to `worker`, and any code changes the
agent made are in the working directory. Nothing is lost — except the agent's
context window, which was going to run out eventually anyway.

Now restart the agent:

```
$ gc start
Starting agent 'worker' (auto-detected: claude)...
Agent 'worker' is running.
```

The fresh agent session starts with zero memory of what came before. But its
startup prompt tells it to check for hooked work. It queries beads:

```
$ gc bead list
ID    STATUS   ASSIGNEE   TITLE
mp-1  closed   worker     Fix the off-by-one error in pagination.go line 42
mp-2  hooked   worker     Refactor the auth middleware to support JWT and session tokens
```

`mp-2` is still hooked to `worker`. The agent sees this, examines the current
state of the codebase (including any partial changes from the previous session),
and continues the refactor from where it left off. No re-explanation needed.
No starting over.

When it finishes:

```
$ gc bead list
ID    STATUS   ASSIGNEE   TITLE
mp-1  closed   worker     Fix the off-by-one error in pagination.go line 42
mp-2  closed   worker     Refactor the auth middleware to support JWT and session tokens
```

> **The aha moment.** The agent ran out of context, started a fresh session,
> and didn't miss a beat — because beads knew what was done and what was left.
> Context windows come and go; beads persist. This is the foundation everything
> else builds on.

---

## Stop the Workspace

When you're done:

```
$ gc stop
Stopping agent 'worker'...
Workspace stopped.
```

---

## What You Learned

This tutorial introduced three of Gas City's five primitives:

| Primitive              | What You Used It For                                                   |
| ---------------------- | ---------------------------------------------------------------------- |
| **Config**             | `hello-world.toml` — declared one agent and the beads backend          |
| **Agent Protocol**     | `gc start` / `gc stop` — started and stopped the agent                 |
| **Task Store (Beads)** | `gc bead create` / `gc bead list` — tracked work through its lifecycle |

The other two primitives (Event Bus and Prompt Templates) aren't needed yet.
They show up when you have multiple agents that need to observe each other
and play different roles. That's Tutorial 03.

**The key insight:** Beads decouple work state from agent state. The agent's
context window is temporary. Beads are permanent. As long as work is tracked
in beads, any agent session — current or future — can pick it up. This is
what makes everything else in Gas City possible: loops, teams, formulas,
health patrol. They all depend on the fact that work state survives agent
restarts.

---

## What's Next

Your agent works one task and stops. If you have a backlog of ten tasks, you'd
have to `gc bead create` each one and `gc start` after each completion. That's
a lot of hand-holding.

In [Tutorial 02 — Looping with Ralph](02-looping-with-ralph.md), you'll add
three lines to your config that turn the agent into a continuous task processor.
It polls for ready beads, claims one, executes it with a clean context, and
loops back for the next. You fill the backlog; the agent drains it.

## Fodder for the next tutorial

$ gc init
Welcome to Gas City SDK!

Example configs available:

> hello-world.toml -- Single agent, single task (simplest)

    ralph.toml       -- Single agent with task loop
    ccat.toml        -- Coordinator + worker pool (Agent Teams)
    gastown.toml     -- Full multi-project orchestration (Gas Town)
    (custom)         -- Start from scratch

Select [hello-world.toml]:

Which coding agent do you use?

> Claude Code

    Codex (OpenAI)
    Gemini CLI
    OpenCode
    Other

Select [Claude Code]:

Created hello-world.toml (Level 1 - Single Agent)
Created .gc/ directory.
Run `gc start` to begin.

````

Let's look at what `gc init` generated:

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
````

Eight lines. That's the entire config. Let's break it down:

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
