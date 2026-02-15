# Gas City Vision

From https://steve-yegge.medium.com/steveys-birthday-blog-34f437139cb5:

- Gas City is an orchestration-builder toolkit (not just an orchestrator).

- GC is the next step beyond Gas Town for folks building their own orchestrators
  from simple to complex in natural.

- GC is a “kit” for building different “town shapes”, i.e. multiple
  architectures / topologies, not one canonical layout. You'll be able to build
  Gas Town in Gas City or any other kind of orchestrator you want, e.g. Ralph,
  Claude Code Agent Teams, via specific configurations of Gas City.

- GC provides a progressive capability model where each level adds a capability.
  The config file grows accordingly. Every level is independently useful.

- GC configuration provides reasonable defaults so that the user only has to
  configure subsystems that they would like to bring online, thus providing for
  a progressive capabilities as needed.

- To populate GC, you can create your own roles, teams, coordination rules, and
  worker instructions (a full configurability surface).

- Roles are expressed in a format external to and not hardcoded into the code,
  e.g. beads, markdown files, etc.

- GC supports wiring in sandboxes, plugins, and hooks (explicit extensibility +
  integration points), etc -- the full range of Gas Town functionality, but
  layered in optionally depending on the user's needs

- GC provides the same set of capabilities and subsystems as GT, but
  configurable instead of hardcoded, e.g. the mayor is part of a "gastown.toml"
  file instead of hardcoded into the code.

- GC builds up higher level concepts on lower level concepts, e.g the mail
  system is built on top of the beads system. The core principles are clearly
  visible and intrinsic whereas the higher level concepts are built up with
  configuration.

- GC is completely transparent and can be traced via the deamon that provides a
  websocket that shows the historical data and streaming changes to data, e.g.
  agent request/response pairs, mail, beads, agent active status, etc.

- Every AI coding agent — regardless of implementation — is accessed through a
  uniform "factory worker" abstraction. This decouples the orchestration logic
  from any specific coding agent (Claude Code, Codex, Gemini, OpenCode, etc) or
  execution substrate (tmux, Agent SDK, custom code). The rest of the
  SDK builds exclusively on this abstraction.
