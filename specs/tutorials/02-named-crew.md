# Tutorial 02 — Named Crew

In Tutorial 01, you created a bead and then manually told a coding agent to
check for ready work. It worked — but *you* were the router. You had to detach
from the mayor, open a new tmux session, start codex, and tell it to look for
beads. That's fine for one bead. It doesn't scale.

What you want is to tell the mayor "here's the work" and have the mayor put it
in front of the right agent. For that, you need two things: **named crew
members** (so the mayor knows who to assign work to) and **hooks** (so the
assignment sticks).

---

## Named Crew

In Tutorial 01, the coding agent's identity was auto-generated:
`tower-of-hanoi-codex`. That name was derived from the rig name and the CLI
process, which is fine for ad-hoc work. But the mayor can't route beads to an
agent that doesn't exist yet.

A **crew member** is a named agent registered on a rig. The mayor knows about
crew members and can assign beads to them. When the crew member starts up, it
finds the work waiting on its hook.

Let's register a crew member on our Tower of Hanoi rig:

```shell
$ cd ~/bright-lights

$ gc crew add tower-of-hanoi --name builder --agent codex
Adding crew member 'builder' to rig 'tower-of-hanoi'...
  Agent type: codex
  Updated AGENTS.md
Crew member added.

$ gc crew list tower-of-hanoi
Crew for rig 'tower-of-hanoi':

  builder:
    Agent: codex
    Status: stopped
```

Now the city knows that `builder` is a codex agent assigned to the Tower of
Hanoi rig. The mayor can refer to it by name.

---

## Hooks

In Tutorial 01, the bead lifecycle was simple: `open → active → closed`. The
coding agent picked up the bead and the status changed to `active`. But there
was no formal assignment — the agent just started working on an open bead.

With named crew, we introduce **hooks** — the mechanism for assigning a bead to
a specific agent. When the mayor hooks a bead to a crew member:

- The bead's status goes from `open` → `hooked`
- The bead's assignee records who it's hooked to
- **Hooking is atomic** — if two agents tried to claim the same bead, exactly
  one would succeed. The other gets a conflict error and moves on. This is how
  Gas City prevents double-work when you scale to many agents.
- When work is done: the bead's status goes `hooked` → `closed`

The full lifecycle is now: `open` → `hooked` → `closed`

> **Why not just `active`?** In Tutorial 01, `active` meant "some agent is
> working on this." But it didn't say *which* agent, and it wasn't atomic.
> `hooked` means "this bead is assigned to a specific crew member, and no one
> else can claim it." That distinction matters as soon as you have more than
> one agent.

---

## The Mayor Routes Work

Now let's see what this looks like in practice. We'll create a few beads and
have the mayor assign them to our `builder` crew member.

```shell
$ gc agent attach mayor

You: Mr. Mayor, I need three things done on the Tower of Hanoi project:
1. Implement the 3-disk solver
2. Add animation to disc moves
3. Write unit tests for the solver

Mayor: I'll create beads for each of those and hook them to builder.

  $ gc bead create "Implement 3-disk solver" --rig tower-of-hanoi
  Created bead: gc-2  (status: open)

  $ gc bead create "Add animation to disc moves" --rig tower-of-hanoi
  Created bead: gc-3  (status: open)

  $ gc bead create "Write unit tests for the solver" --rig tower-of-hanoi
  Created bead: gc-4  (status: open)

  $ gc bead hook gc-2 --assignee builder
  Hooked bead gc-2 to 'builder' (status: open → hooked)

  $ gc bead hook gc-3 --assignee builder
  Hooked bead gc-3 to 'builder' (status: open → hooked)

  $ gc bead hook gc-4 --assignee builder
  Hooked bead gc-4 to 'builder' (status: open → hooked)

Done. Three beads created and assigned to builder. Here's the backlog:

  $ gc bead ready
  ID    STATUS   ASSIGNEE   TITLE
  gc-2  hooked   builder    Implement 3-disk solver
  gc-3  hooked   builder    Add animation to disc moves
  gc-4  hooked   builder    Write unit tests for the solver
```

The mayor created the beads and hooked them to `builder`. The beads are now
assigned — they're waiting for builder to pick them up.

> **Notice:** The mayor didn't run codex. The mayor didn't write code. The
> mayor created beads and assigned them. That's the overseer's job.

---

## Starting the Crew

Now start the builder. Because `builder` is a registered crew member, you can
start it through `gc`:

```shell
$ gc crew start builder --rig tower-of-hanoi
Starting crew member 'builder' (codex) in tmux session 'bright-lights/builder'...
  Working directory: ~/projects/tower-of-hanoi
  Checking hook... found 3 beads assigned to 'builder'
  Starting with gc-2: "Implement 3-disk solver"
Builder is running.
```

The crew member starts up, checks its hook, and finds gc-2 waiting. It gets to
work immediately — no nudging required.

You can watch it work:

```shell
$ gc agent attach builder
```

Or check progress from another terminal:

```shell
$ gc bead list
ID    STATUS   ASSIGNEE                TITLE
gc-1  closed   tower-of-hanoi-codex    Build a Tower of Hanoi app
gc-2  active   builder                 Implement 3-disk solver
gc-3  hooked   builder                 Add animation to disc moves
gc-4  hooked   builder                 Write unit tests for the solver
```

Builder is working on gc-2. When it finishes, it closes gc-2 and you can
tell it to pick up gc-3:

```shell
$ gc bead list
ID    STATUS   ASSIGNEE                TITLE
gc-1  closed   tower-of-hanoi-codex    Build a Tower of Hanoi app
gc-2  closed   builder                 Implement 3-disk solver
gc-3  hooked   builder                 Add animation to disc moves
gc-4  hooked   builder                 Write unit tests for the solver
```

---

## What You Learned

This tutorial added two concepts on top of Tutorial 01:

| Concept          | What It Does                                                       |
| ---------------- | ------------------------------------------------------------------ |
| **Named Crew**   | `gc crew add` — registered agents on rigs with stable identities   |
| **Hooks**        | `gc bead hook` — atomic bead assignment to a specific crew member  |

The mayor can now create beads and route them to named crew members. The crew
members find their assigned work on startup. No manual nudging needed — just
start the crew and the work is waiting.

But you're still starting the crew manually and telling it to pick up the next
bead after each one finishes. Wouldn't it be nice if the crew member just...
kept going? Clear its context, check the hook, pick up the next bead, repeat?

That's [Tutorial 03 — The Ralph Loop](03-ralph-loop.md).

---

## Spec Changes Needed

> This section tracks DX decisions from the tutorial that need to flow back
> into `gas-city-spec.md`. Don't delete until the spec is updated.

- **`gc crew add <rig> --name <name> --agent <type>`** — registers a named
  crew member on a rig. New command not in spec.

- **`gc crew list <rig>`** — lists crew members for a rig. New command.

- **`gc crew start <name> --rig <rig>`** — starts a crew member in a tmux
  session, auto-checks hook on startup. New command.

- **`gc bead hook <bead-id> --assignee <name>`** — assigns a bead to a crew
  member. Transitions status from `open` → `hooked`. New command.

- **Bead lifecycle update.** Tutorial 01 used `open → active → closed`.
  Tutorial 02 introduces `open → hooked → closed` for assigned work. Need to
  reconcile: is `active` still a valid status for ad-hoc (unnamed) agents, or
  does everything go through `hooked` now?

- **`gc crew add` updates AGENTS.md.** Registering a crew member updates the
  rig's rules files so the agent knows its identity and role.

- **Crew member auto-starts with hooked work.** When a crew member starts via
  `gc crew start`, it automatically checks for hooked beads and begins working
  on the first one. This startup behavior needs to be in the spec.

- **Hook ordering.** When multiple beads are hooked to the same crew member,
  what order are they worked? FIFO by hook time? Priority field? Need to
  decide.

- **Relationship between `gc agent attach` and `gc crew start`.** The mayor
  is attached via `gc agent attach mayor`. Crew members are started via
  `gc crew start builder`. Are these the same mechanism with different names,
  or genuinely different? The mayor is always running; crew members are started
  on demand.
