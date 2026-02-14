# Gas City Vision

From https://steve-yegge.medium.com/steveys-birthday-blog-34f437139cb5:

- It’s an orchestration-builder toolkit (not just an orchestrator)

- It’s positioned as the next step beyond Gas Town for “Level 8” folks building
  their own orchestrators

- It’s a “kit” for building different “town shapes”, i.e. multiple architectures
  / topologies, not one canonical layout. You'll be able to build Gas Town in
  Gas City or any other kind of orchestrator you want, e.g. Ralph, Claude Code
  Agent Teams, via specific configurations of Gas City.

- You can create your own roles, teams, coordination rules, and worker
  instructions (a full configurability surface)

- Roles can be expressed in a dynamic format external to and not hardcoded into
  the code

- GC supports wiring in sandboxes, plugins, and hooks (explicit extensibility +
  integration points)

- GC provides a progressive capability model where each level adds a capability.
  The config file grows accordingly. Every level is independently useful.

- every agent — regardless of implementation — is accessed through a uniform
  "factory worker" abstraction. This decouples the orchestration logic from any
  specific coding agent (Claude Code, Codex, Gemini, OpenCode, etc) or execution
  substrate (tmux, Docker, Agent SDK, custom code). The rest of the SDK builds
  exclusively on this abstraction.
