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
To configure your new city, add a `settings.yaml` file.

To get started with one of the built-in configurations, use `gc init`.

To add a rig (project), [TODO]

[TODO: the rest]
```

Starting a city uses the configuration to ensure that you have the agents you
need to do your work. You can update the configuration at any time, stop and
restart your city for a new configurations to take affect. If you specify no
configuration, you'll get the default which is what we'll use for the rest of
this tutorial.

## Adding a project

To associate a project (called a "rig") with a city, you add it:

```shell
$ cd ~/bright-lights

$ gc rig add ~/projects/tower-of-hanoi
[TODO: "adding rig..."]

$ gc rig list

Rigs in /Users/csells/bright-lights:

  tower-of-hanoi:
    Agents: [mayor]
```

The "gc rig" command needs to know which city you're talking about. The easiest
way to specify that is to be in the configuration folder of the city itself, but
you can also specify the city manually via the `--city` argment.

Because we're getting the default GC configuration, we have only a single agent
-- the mayor -- who is wwho you'll talk to do the planning and work.

## Create a Task

To give an agent something to do, create a bead that represents a task. You can
do that manually or use the mayor to do that work. Creating a bead manually
looks like this:

```shell
# create the bead manually
$ gc bead create "build a Tower of Hanoi app"
Created bead: gc-1  (status: open)

# list the open beads (use this less)
$ gc bead list

[TODO: show command output]

# list the beads ready to work on (use this more)
$ gc bead ready

[TODO: show command output]
```

However, I recommend you talk to the mayor about the work you'd like to do
instead:

```shell
$ gc agent attach mayor
Attaching to agent 'mayor' (tmux session: bright-lights/mayor)...

╭────────────────────────────────────────╮
│ ✻ Welcome to Claude Code!              │
│   /help for help                       │
│                                        │
│   cwd: ~/projects/tower-of-hanoi       │
╰────────────────────────────────────────╯

You: Mr. Mayor, can you create a bead to build a Tower of Hanoi app? thanks!

Mayor: Sure! I'll create that bead for you.

  $ gc bead create "Build a Tower of Hanoi app"
  Created bead: gc-1  (status: open)

Done — gc-1 is ready to go. Want me to start working on it?

You: Can you list the ready beads?

Mayor: Of course.

  $ gc bead ready
  ID    STATUS   TITLE
  gc-1  open     Build a Tower of Hanoi app

Just the one bead in the backlog right now. Would you like me to get to work on
it?

You: [switch to another shell instance]
```

The act of "attaching" to the mayor via `gt agent attach` brings up the single
instance of that agent running in a tmux session.

A new bead starts with a status of `open` — available for claiming. No assignee
yet.

```
$ gc bead show gc-1
ID:       gc-1
Status:   open
Type:     task
Title:    Build a Tower of Hanoi app
Rig:      tower-of-hanoi
Created:  2026-02-16 10:30:00
Assignee: —
```

> **Two things to notice:**
>
> 1. The bead has an ID (`gc-1`). Every bead in this city gets a unique ID.
> 2. The bead is stored on disk — not in the agent's context window. Agents come
>    and go. Beads persist.

---

## Let's get to work!

Now let's use a CLI coding agent to pick up that work for our rig. Leave the
mayor for other work and start a new shell instance. Now `cd` to your project
directory so we can put an agent to work there.

Because you've added the rig to the "bright-lights" city, any agent that
supports the AGENTS.md (most of them) or CLAUDE.md (Claude Code) rules files has
already been configured with the information it needs to understand tasks
expressed as beads:

```shell
$ cd ~/projects/tower-of-hanoi
$ codex # or claude or gemini or ...

[TODO: agent output]

You: can you check what beads are ready?

[TODO: agent output asking if you want to do this work]

You: yes, please!

[TODO: agent output]
```

You can watch it build your app right in the tmux session, or detach from the
tmux sessions (`Ctrl-b d`) and let it cook.

Check the bead status from another terminal any time you like:

```shell
$ gc bead list
ID    STATUS   ASSIGNEE   TITLE
gc-1  active   mayor      Build a Tower of Hanoi app
```

When the agent finishes, it closes the bead:

```shell
$ gc bead list
ID    STATUS   ASSIGNEE   TITLE
gc-1  closed   mayor      Build a Tower of Hanoi app
```

That's it. The coding agent has now built your app. The bead records that the
work happened, who did it, and when it closed.

---

## What You Learned

This tutorial used three of Gas City's five primitives:

| Primitive              | What You Used It For                                           |
| ---------------------- | -------------------------------------------------------------- |
| **Config**             | Default city configuration — one mayor, beads backend          |
| **Agent Protocol**     | `gc start` / `gc stop` / `gc agent attach` — managed the mayor |
| **Task Store (Beads)** | `gc bead create` / `gc bead list` — tracked the work           |

The other two primitives (Event Bus and Prompt Templates) aren't needed yet.
They show up when you have multiple agents that need to observe each other and
play different roles. That's [Tutorial 03](03-agent-team.md).

---

## What's Next

At this point, you've got yourself a working orchestration system. You can use
the mayor to create beads and pull in your beads from your working agents on
demand.

In Tutorial 02, we'll see how to manage multiple rigs and route work from the
mayor to multiple rigs via named "crew" (agents assigned to your rigs).

[TODO: push the Ralph loop ahead one step to make room for the multi-project
tutorial]
