# Gas City Vision

From https://steve-yegge.medium.com/steveys-birthday-blog-34f437139cb5:

- Gas City is an orchestration-builder toolkit (not just an orchestrator)

- GC is the next step beyond Gas Town for “Level 8” folks building their own
  orchestrators

- GC is a “kit” for building different “town shapes”, i.e. multiple
  architectures / topologies, not one canonical layout. You'll be able to build
  Gas Town in Gas City or any other kind of orchestrator you want, e.g. Ralph,
  Claude Code Agent Teams, via specific configurations of Gas City.

- To populate GC, you can create your own roles, teams, coordination rules, and
  worker instructions (a full configurability surface)

- Roles are expressed in a format external to and not hardcoded into the code,
  e.g. beads, markdown files, etc.

- GC supports wiring in sandboxes, plugins, and hooks (explicit extensibility +
  integration points)

- GC provides a progressive capability model where each level adds a capability.
  The config file grows accordingly. Every level is independently useful.

- Every agent — regardless of implementation — is accessed through a uniform
  "factory worker" abstraction. This decouples the orchestration logic from any
  specific coding agent (Claude Code, Codex, Gemini, OpenCode, etc) or execution
  substrate (tmux, Docker, Agent SDK, custom code). The rest of the SDK builds
  exclusively on this abstraction.
