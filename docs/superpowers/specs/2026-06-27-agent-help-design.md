# Agent Help Design

## Summary

Replace the current `--help-llm` feature with `--help-agent`: a generic,
side-effect-free Markdown context document for Codex-style agents operating a
CLI through normal computer use.

The feature belongs in yargs as a generic library capability. Downstream CLIs
such as yeet should be able to opt into richer output through yargs metadata,
without yargs knowing the downstream domain and without writing files such as
repo-local skills.

## Goals

- Make `app --help` point agents toward `app --help-agent`.
- Generate agent-readable context from yargs metadata.
- Make the output useful when loaded directly into an agent context.
- Support global, group, flat command, and grouped command help.
- Remove the old `--help-llm` API and output style.
- Keep normal human help and parsing behavior intact.

## Non-Goals

- No filesystem writes or generated skill files.
- No yeet-specific behavior in yargs.
- No JSON output in the first version.
- No compatibility layer for `--help-llm`.

## Public Surface

Human help should advertise the agent help path:

```text
--help-agent    Show agent-readable CLI context
```

Supported invocations:

```bash
app --help-agent
app COMMAND --help-agent
app GROUP --help-agent
app GROUP COMMAND --help-agent
```

`app help COMMAND` remains human help unless `--help-agent` is explicitly
present.

`--help-agent` is successful help output. Parse-and-handle APIs should treat it
like normal help and return `ErrShown` after printing. Lower-level parse APIs
should return a new sentinel, `ErrHelpAgent`.

Remove `ErrHelpLLM`, `Generate*HelpLLM`, and `--help-llm` dispatch handling.

## Agent Metadata

Replace `LLMInstructions` with generic optional agent metadata.

```go
type AgentInfo struct {
    Summary   string
    Rules     []string
    Safety    []string
    Discovery []string
}
```

Add `Agent AgentInfo` to:

- `CommandInfo`
- `SubCommandInfo`
- `GroupInfo`

All fields are optional. If a CLI provides no `AgentInfo`, yargs still emits
useful context from descriptions, aliases, examples, flag tags, positional
tags, defaults, visibility, and command structure.

Field intent:

- `Summary`: concise agent-oriented purpose text.
- `Rules`: generic or domain-specific operating guidance.
- `Safety`: commands or options that need extra care.
- `Discovery`: how agents should inspect deeper command-specific context.

## Internal Command Model

Unify agent help around a resolved target instead of maintaining independent
global, group, subcommand, and grouped-command generators that can drift.

```go
type AgentHelpTargetKind string

const (
    AgentHelpGlobal       AgentHelpTargetKind = "global"
    AgentHelpGroup        AgentHelpTargetKind = "group"
    AgentHelpCommand      AgentHelpTargetKind = "command"
    AgentHelpGroupCommand AgentHelpTargetKind = "group_command"
)

type AgentHelpTarget struct {
    Kind       AgentHelpTargetKind
    Path       []string
    Command    CommandInfo
    Group      GroupInfo
    SubCommand SubCommandInfo
    FlagsSchema any
    ArgsSchema  any
}
```

Extend `CommandSpec` with command flag metadata:

```go
type CommandSpec struct {
    Info        SubCommandInfo
    FlagsSchema any
    ArgsSchema  any
}
```

`Registry` becomes the preferred source for detailed agent help because it can
carry the command tree, command metadata, flag schemas, and positional schemas.
`HelpConfig` can still generate basic agent help, but it cannot show detailed
flags or args that were never provided.

## Generation APIs

Provide a small replacement API surface:

```go
func GenerateAgentHelp[G any](config HelpConfig, globalFlagsExample G) string

func GenerateAgentHelpForCommand[G any, S any, A any](
    config HelpConfig,
    commandPath []string,
    globalFlagsExample G,
    flagsExample S,
    argsExample A,
) string

func GenerateAgentHelpFromRegistry[G any](
    reg Registry,
    commandPath []string,
    globalFlagsExample G,
) string
```

`GenerateAgentHelpFromRegistry` should resolve aliases, groups, and grouped
commands from the registry-derived `HelpConfig`, then attach `FlagsSchema` and
`ArgsSchema` from the matching `CommandSpec`.

## Markdown Output

The Markdown should read like an ephemeral skill, not API documentation. Avoid
the phrase "LLM Instructions"; use actionable sections.

Global output:

```text
# app Agent Context

## Purpose
## Operating Rules
## Discovery
## Global Options
## Commands
## Command Groups
## Examples
```

Group output:

```text
# app GROUP Agent Context

## Purpose
## Operating Rules
## Discovery
## Commands
## Global Options
## Examples
```

Flat command and grouped-command output:

```text
# app COMMAND Agent Context

## Purpose
## Usage
## Arguments
## Options
## Global Options
## Examples
## Safety Notes
```

Grouped-command headings should include the full command path, for example:

```text
# app GROUP COMMAND Agent Context
```

Default generated operating rules should include:

- Prefer exact examples when they match the task.
- Use command-specific agent help before running an unfamiliar command.
- Do not invent flags; use only flags listed in this context or command help.
- Preserve arguments after `--` as payload or application arguments.

Yargs should merge these defaults with app-provided `AgentInfo.Rules`,
`AgentInfo.Safety`, and `AgentInfo.Discovery`.

## Dispatch Behavior

- `app --help-agent` prints global agent context.
- `app COMMAND --help-agent` prints detailed flat command context.
- `app GROUP --help-agent` prints group agent context.
- `app GROUP COMMAND --help-agent` prints detailed grouped command context.
- Aliases are resolved before choosing the target.
- Canonical command names are used in headings and examples.
- Aliases remain documented in the relevant command section.
- Hidden commands and groups remain hidden.

Unknown targets should produce short agent-oriented errors with a valid
discovery command:

```text
Unknown command: deploy. Run 'app --help-agent' to inspect available commands.
```

## Error Handling

- Add `ErrHelpAgent`.
- Remove `ErrHelpLLM`.
- Parse-and-handle helpers print agent help and return `ErrShown`.
- Dispatcher helpers print agent help and return nil.
- Old `--help-llm` becomes an unknown flag or unknown command path after
  removal.

## Testing

Cover these behaviors:

- Human `--help` advertises `--help-agent`.
- Human `--help` no longer advertises `--help-llm`.
- Global agent help includes purpose, discovery, visible commands, visible
  groups, aliases, examples, global options, and generic rules.
- Hidden commands and groups are excluded.
- Flat command agent help includes usage, positional args, command flags,
  global flags, defaults, aliases, examples, and safety notes.
- Group help lists visible child commands and discovery commands.
- Group command agent help includes `FlagsSchema` and `ArgsSchema` from
  `Registry`.
- Alias invocation such as `app rm --help-agent` resolves to canonical command
  help while documenting `rm` as an alias.
- `--help-agent` returns `ErrHelpAgent` from lower-level parse APIs and
  `ErrShown` from parse-and-handle APIs.
- Existing normal help, parsing, aliases, registry resolution, and dispatch
  tests keep passing.

## Migration Notes

Downstream CLIs should replace:

- `--help-llm` with `--help-agent`
- `ErrHelpLLM` with `ErrHelpAgent`
- `LLMInstructions` with `AgentInfo`
- `Generate*HelpLLM` calls with the new agent help generators

For CLIs that already maintain command schemas, prefer moving those schemas into
`Registry` so grouped command agent help can include both flags and positional
arguments without custom interception code.
