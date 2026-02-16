# Gas City Tutorials — Series Overview

If you're reading this, you've probably tried AI coding agents on your projects.
Perhaps they worked great or perhaps not. Most likely they didn't do quite what
you wanted. Perhaps you've tried a Ralph loop or a Claude Code Agent Team or Gas
Town. Maybe it was amazing or maybe there was a lot of hand-holding to get the
agents to do what you wanted in the order and using the specific
process/tools/details that you specify.

If any of that (or even all of it) is true, then what you need isn't just an
off-the-shelf agentic coding orchestration tool. What you need is a ODK --
Orchestration Development Kit. Luckily, you've found one!

Gas City is the generalization and modularization of the popular Gas Town
orchestration tool so that you can build it to work exactly how you want it to.
Want to use the industry-leading Beads for task management so that your context
lives beyond any specific chat session? We got you. Perhaps you'd like to split
the work between Claude Code, Gemini, Codex, Amp, OpenCode, etc. to ensure
you're using the best agent for the job or just getting multiple perspectives.
Done. Or maybe you want the full-bore Gas Town experience with your own
customizations that Gas Town just doesn't support. Here for that too.

Gas City is an orchestration-builder SDK — an ODK for composing multi-agent
coding workflows. You don't get a fixed orchestrator. You get the building
blocks to construct your own: from a single agent that survives context loss,
all the way to a self-healing team of eight roles across multiple repos.

These tutorials build Gas City up from the core principles baked into Gas City.
Each one adds a new capability to your config. At the end of each tutorial,
you'll have a working orchestrator. Want to add more features to get it just
right? Then keep on reading with the next tutorial. By the time you get to the
end, you'll have the entire Gas Town stack built - and you'll understand every
line because you added them all yourself.

Note: You get configurations for "hello, world", "Ralph loop", "Agent Teams" and
"Gas Town" out of the box with Gas City, so you don't have to rebuild them; you
can just use them.

---

## What You'll Build

Each tutorial solves a real problem that plain coding agents can't handle on
their own. The config grows; it never resets.

| Tutorial                                              | Problem                                                                                        | What You Add                                   |
| ----------------------------------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| [01 — Hello, Gas City](01-hello-gas-city.md)          | "My agent ran out of context and forgot everything."                                           | Agent + Task Store + Config                    |
| [02 — Looping with Ralph](02-looping-with-ralph.md)   | "Implement each task with a clean context; don't stop until you've met my definition of done." | `[agents.loop]`                                |
| [03 — Agent Team](03-agent-team.md)                   | "One agent is too slow!"                                                                       | Multiple agents + roles + messaging + dispatch |
| [04a — Formulas](04a-formulas.md)                     | "I want to provide a specific workflow for my agent team."                                     | Formulas + molecules                           |
| [04b — Health Patrol](04b-health-patrol.md)           | "Keep your agents going without babysitting them."                                             | Daemon + health monitoring                     |
| [04c — Plugins](04c-plugins.md)                       | "I want automated maintenance, not manual chores."                                             | Plugin system + gate conditions                |
| [04d — Full Orchestration](04d-full-orchestration.md) | "I want the full Gas Town!"                                                                    | Multi-project + all roles                      |

Tutorial 01 gives you a working system. Everything after that makes it
more capable.

---

## Prerequisites

- **git tooling and a repo** to work in (the tutorials use a sample project)
- **An AI coding agent** available — Claude Code (recommended), Codex, or Gemini
  CLI

No prior Gas Town experience required. These tutorials start from zero.

---

### Conventions

**Shell commands** are prefixed with `$`:

```
$ gc start
Starting agent 'worker'...
Agent 'worker' is running.
```

**TOML config** is shown as file contents with the filename above:

```toml
# workspace.toml
[workspace]
name = "my-project"
```

**Explanatory callouts** highlight key insights:

> **What just happened?** The agent ran out of context, started a fresh
> session, and didn't miss a beat — because beads knew what was done and
> what was left.

Each tutorial builds on the previous config. When a tutorial says "add this
section," you're adding to your existing file, not starting from scratch.

---

## The Building Blocks — A Quick Vocabulary

Gas City's design rests on **five primitives** and **four mechanisms**. The
primitives are infrastructure the SDK provides. The mechanisms are built from
primitives via your configuration. You don't need to memorize these now —
each tutorial introduces them when you need them.

### Five Primitives

| Primitive              | What It Does                                                                                                     |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------- |
| **Agent Protocol**     | Start, stop, prompt, and observe agents — regardless of which AI powers them                                     |
| **Task Store (Beads)** | Track work units with status, dependencies, and atomic claiming                                                  |
| **Event Bus**          | Publish/subscribe log of all system activity                                                                     |
| **Config**             | TOML files that declare what exists and how it behaves — the SDK activates subsystems based on what's configured |
| **Prompt Templates**   | Markdown templates that define what each role does — the behavioral specification                                |

### Four Mechanisms (Built from Primitives)

| Mechanism                | What It Does                                                               | Built From                            |
| ------------------------ | -------------------------------------------------------------------------- | ------------------------------------- |
| **Messaging**            | Inter-agent communication — persistent mail and immediate nudges           | Task Store + Agent Protocol           |
| **Formulas & Molecules** | Reusable multi-step workflows — define once, instantiate per task          | Config + Task Store + Templates       |
| **Dispatch (Sling)**     | One-command work assignment — spawn agent, create workflow, assign, notify | All primitives + Messaging + Formulas |
| **Health Patrol**        | Self-healing — detect stalls, restart agents, recover work                 | Agent Protocol + Event Bus + Config   |

The five primitives have no circular dependencies. The four mechanisms
introduce no new irreducible ideas — they're compositions of the primitives.
This means the system is as simple as it can be while still handling
everything from a solo agent to a multi-repo team.

For the full conceptual treatment, see the
[Gas City Concepts](../gas-city-concepts.md) document.

---

## Where Each Tutorial Fits

The tutorials map to Gas City's progressive capability model. The SDK
activates subsystems based on which config sections you've written.
Early tutorials use few sections; later ones use more.

```
Tutorial 01 — Hello, Gas City
  Primitives:  Agent Protocol, Task Store, Config
  Mechanisms:  (none)
  You get:     One agent, tracked work, context survival

Tutorial 02 — Looping with Ralph
  Adds:        [agents.loop] config
  You get:     Continuous task processing from a backlog

Tutorial 03 — Agent Team
  Adds:        Event Bus, Prompt Templates, Messaging, Dispatch
  You get:     Coordinator + workers, typed roles, work distribution

Tutorial 04a — Formulas
  Adds:        Formulas & Molecules
  You get:     Reusable multi-step workflows

Tutorial 04b — Health Patrol
  Adds:        Daemon, Health Patrol
  You get:     Self-healing, stall detection, automatic recovery

Tutorial 04c — Plugins
  Adds:        Plugin system with gate conditions
  You get:     Automated maintenance tasks on schedule/event triggers

Tutorial 04d — Full Orchestration
  Adds:        Multi-project config, all 8 roles
  You get:     The complete Gas Town — rebuilt from config
```

By tutorial 04d, your config file defines the full Gas Town orchestration
system. Every line in it exists because you added it to solve a specific
problem.

---

## Next Step

Ready? Start with [Tutorial 01 — Hello, Gas City](01-hello-gas-city.md).
