# Agent Help Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace yargs `--help-llm` with generic Markdown `--help-agent` output that Codex-style agents can load as ephemeral CLI context.

**Architecture:** Keep human help generation intact, but replace the old LLM-specific generator and sentinel with an agent-help path. Add generic `AgentInfo` metadata and `FlagsSchema` on `CommandSpec`, then generate global, group, flat-command, and grouped-command agent help through a shared resolved-target model.

**Tech Stack:** Go standard library, yargs reflection-based flag and positional schema extraction, Go `testing`, `mise` tasks.

---

## Current Workspace Notes

- The repository currently has unstaged edits in `yargs.go` and `yargs_test.go` from earlier alias-help work. Treat them as user work and preserve them.
- The design spec is committed at `docs/superpowers/specs/2026-06-27-agent-help-design.md`.
- The implementation should stay generic. Do not import or reference yeet.

## File Structure

- Modify `yargs.go`: constants, sentinel errors, public metadata structs, registry schema, agent-help target resolution, Markdown rendering, parse helpers, and dispatcher help handling.
- Rename `yargs_llm_test.go` to `yargs_agent_test.go`: replace LLM help tests with agent help tests.
- Modify `yargs_test.go`: update human help assertions and dispatch tests that mention `--help-llm`.
- Modify `README.md`: replace LLM help documentation with agent help documentation.
- Modify `doc.go`: update package docs for `AgentInfo`, `--help-agent`, and registry `FlagsSchema`.

---

### Task 1: Add Agent Metadata Tests

**Files:**
- Create: `yargs_agent_test.go`
- Delete: `yargs_llm_test.go`
- Modify: `yargs_test.go`

- [ ] **Step 1: Rename the old LLM test file**

Run:

```bash
git mv yargs_llm_test.go yargs_agent_test.go
```

Expected: `git status --short` shows `R  yargs_llm_test.go -> yargs_agent_test.go` plus the existing unstaged files.

- [ ] **Step 2: Replace old LLM tests with failing agent tests**

Replace the contents of `yargs_agent_test.go` with:

```go
// Copyright (c) 2025 AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yargs

import (
	"strings"
	"testing"
)

func TestGenerateAgentHelpGlobal(t *testing.T) {
	type GlobalFlags struct {
		Verbose bool `flag:"verbose" short:"v" help:"Enable verbose logging"`
		Plain   bool `flag:"plain" help:"Use plain output"`
	}

	config := HelpConfig{
		Command: CommandInfo{
			Name:        "testcli",
			Description: "Test CLI application",
			Agent: AgentInfo{
				Summary:   "Use testcli to operate test services.",
				Rules:     []string{"Prefer --plain when parsing output in automation."},
				Safety:    []string{"Do not run destructive commands unless the user asked for them."},
				Discovery: []string{"Run testcli COMMAND --help-agent before using an unfamiliar command."},
			},
			Examples: []string{
				"testcli status",
				"testcli run ./app",
			},
		},
		SubCommands: map[string]SubCommandInfo{
			"status": {
				Name:        "status",
				Description: "Show status",
				Agent: AgentInfo{
					Summary: "Use for read-only service status checks.",
				},
				Examples: []string{
					"testcli status",
					"testcli status --format=json",
				},
			},
			"remove": {
				Name:        "remove",
				Description: "Remove a service",
				Aliases:     []string{"rm"},
				Agent: AgentInfo{
					Safety: []string{"This command can remove running service state."},
				},
			},
			"hidden": {
				Name:        "hidden",
				Description: "Hidden command",
				Hidden:      true,
			},
		},
		Groups: map[string]GroupInfo{
			"docker": {
				Name:        "docker",
				Description: "Docker commands",
				Agent: AgentInfo{
					Summary: "Use docker commands for compose image maintenance.",
				},
				Commands: map[string]SubCommandInfo{
					"run": {
						Name:        "run",
						Description: "Run a container",
					},
					"hidden": {
						Name:        "hidden",
						Description: "Hidden group command",
						Hidden:      true,
					},
				},
			},
			"hiddengroup": {
				Name:        "hiddengroup",
				Description: "Hidden group",
				Hidden:      true,
			},
		},
	}

	output := GenerateAgentHelp(config, GlobalFlags{})

	mustContain(t, output,
		"# testcli Agent Context",
		"## Purpose",
		"Use testcli to operate test services.",
		"## Operating Rules",
		"Prefer exact examples when they match the task.",
		"Do not invent flags; use only flags listed in this context or command help.",
		"Prefer --plain when parsing output in automation.",
		"## Discovery",
		"Run `testcli status --help-agent` for command-specific context.",
		"Run `testcli docker --help-agent` for group-specific context.",
		"## Global Options",
		"### `--verbose`",
		"short: `-v`",
		"Enable verbose logging",
		"## Commands",
		"### `remove`",
		"**Aliases**: `rm`",
		"## Command Groups",
		"### `docker`",
		"## Examples",
		"testcli status",
		"## Safety Notes",
		"Do not run destructive commands unless the user asked for them.",
	)

	mustNotContain(t, output,
		"LLM Instructions",
		"--help-llm",
		"### `hidden`",
		"### `hiddengroup`",
	)
}

func TestGenerateAgentHelpForFlatCommand(t *testing.T) {
	type GlobalFlags struct {
		Verbose bool `flag:"verbose" short:"v" help:"Enable verbose logging"`
	}
	type RunFlags struct {
		Service string `flag:"service" help:"Service name"`
		Port    int    `flag:"port" short:"p" help:"Port number" default:"8080"`
	}
	type RunArgs struct {
		Payload string `pos:"0" help:"Path to payload"`
	}

	config := HelpConfig{
		Command: CommandInfo{Name: "testcli"},
		SubCommands: map[string]SubCommandInfo{
			"run": {
				Name:        "run",
				Description: "Deploy and run a service",
				Agent: AgentInfo{
					Summary: "Use when the user wants to deploy a payload.",
					Safety:  []string{"This command changes service runtime state."},
				},
				Examples: []string{
					"testcli run ./app",
					"testcli run --service=api ./app",
				},
			},
		},
	}

	output := GenerateAgentHelpForCommand(config, []string{"run"}, GlobalFlags{}, RunFlags{}, RunArgs{})

	mustContain(t, output,
		"# testcli run Agent Context",
		"Use when the user wants to deploy a payload.",
		"## Usage",
		"testcli [GLOBAL_OPTIONS] run <PAYLOAD> [OPTIONS]",
		"## Arguments",
		"### `PAYLOAD`",
		"Path to payload",
		"- **Required**: true",
		"## Options",
		"### `--service`",
		"### `--port`",
		"short: `-p`",
		"- **Default**: `8080`",
		"## Global Options",
		"### `--verbose`",
		"## Examples",
		"testcli run ./app",
		"## Safety Notes",
		"This command changes service runtime state.",
	)
}

func TestGenerateAgentHelpForGroupAndGroupCommandFromRegistry(t *testing.T) {
	type GlobalFlags struct {
		Verbose bool `flag:"verbose" short:"v" help:"Enable verbose logging"`
	}
	type DockerRunFlags struct {
		Detach bool `flag:"detach" short:"d" help:"Run in background"`
	}
	type DockerRunArgs struct {
		Image string `pos:"0" help:"Image reference"`
	}

	reg := Registry{
		Command: CommandInfo{Name: "testcli"},
		Groups: map[string]GroupSpec{
			"docker": {
				Info: GroupInfo{
					Name:        "docker",
					Description: "Docker commands",
					Agent: AgentInfo{
						Discovery: []string{"Inspect the specific docker command before running it."},
					},
				},
				Commands: map[string]CommandSpec{
					"run": {
						Info: SubCommandInfo{
							Name:        "run",
							Description: "Run a container",
							Aliases:     []string{"start"},
							Examples:    []string{"testcli docker run nginx"},
						},
						FlagsSchema: DockerRunFlags{},
						ArgsSchema:  DockerRunArgs{},
					},
				},
			},
		},
	}

	groupOutput := GenerateAgentHelpFromRegistry(reg, []string{"docker"}, GlobalFlags{})
	mustContain(t, groupOutput,
		"# testcli docker Agent Context",
		"## Commands",
		"### `docker run`",
		"Run `testcli docker run --help-agent` for command-specific context.",
		"Inspect the specific docker command before running it.",
	)

	commandOutput := GenerateAgentHelpFromRegistry(reg, []string{"docker", "start"}, GlobalFlags{})
	mustContain(t, commandOutput,
		"# testcli docker run Agent Context",
		"**Aliases**: `start`",
		"## Arguments",
		"### `IMAGE`",
		"Image reference",
		"## Options",
		"### `--detach`",
		"short: `-d`",
		"Run in background",
		"testcli docker run nginx",
	)
}

func TestGenerateAgentHelpUnknownTargets(t *testing.T) {
	config := HelpConfig{
		Command: CommandInfo{Name: "testcli"},
		SubCommands: map[string]SubCommandInfo{
			"status": {Name: "status", Description: "Show status"},
		},
	}

	output := GenerateAgentHelpForCommand(config, []string{"deploy"}, struct{}{}, struct{}{}, struct{}{})
	mustContain(t, output, "Unknown command: deploy.", "Run `testcli --help-agent` to inspect available commands.")
}

func TestVariadicArgsInAgentHelp(t *testing.T) {
	type StartArgs struct {
		Services []string `pos:"0+" help:"Service names to start"`
	}
	config := HelpConfig{
		Command: CommandInfo{Name: "testcli"},
		SubCommands: map[string]SubCommandInfo{
			"start": {Name: "start", Description: "Start services"},
		},
	}

	output := GenerateAgentHelpForCommand(config, []string{"start"}, struct{}{}, struct{}{}, StartArgs{})
	mustContain(t, output, "<SERVICES...>", "- **Variadic**: true (minimum: 1)")
}

func mustContain(t *testing.T, output string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(output, value) {
			t.Fatalf("output missing %q:\n%s", value, output)
		}
	}
}

func mustNotContain(t *testing.T, output string, values ...string) {
	t.Helper()
	for _, value := range values {
		if strings.Contains(output, value) {
			t.Fatalf("output unexpectedly contains %q:\n%s", value, output)
		}
	}
}
```

- [ ] **Step 3: Add human help expectation tests**

In `yargs_test.go`, update the existing help-generation assertions or add this test near the other help tests:

```go
func TestGenerateGlobalHelpAdvertisesAgentHelp(t *testing.T) {
	type GlobalFlags struct {
		Verbose bool `flag:"verbose" short:"v" help:"Verbose mode"`
	}
	config := HelpConfig{
		Command: CommandInfo{Name: "myapp", Description: "My app"},
		SubCommands: map[string]SubCommandInfo{
			"run": {Name: "run", Description: "Run a thing"},
		},
	}

	helpText := GenerateGlobalHelp(config, GlobalFlags{})
	if !strings.Contains(helpText, "--help-agent") {
		t.Fatalf("human help should advertise --help-agent:\n%s", helpText)
	}
	if strings.Contains(helpText, "--help-llm") {
		t.Fatalf("human help should not advertise --help-llm:\n%s", helpText)
	}
}
```

- [ ] **Step 4: Run focused tests to verify failure**

Run:

```bash
go test ./... -run 'TestGenerateAgentHelp|TestGenerateGlobalHelpAdvertisesAgentHelp' -count=1
```

Expected: FAIL with errors including `undefined: AgentInfo`, `undefined: GenerateAgentHelp`, or a human help assertion failure.

- [ ] **Step 5: Commit the failing tests**

Run:

```bash
git add yargs_agent_test.go yargs_llm_test.go yargs_test.go
git commit -m "test: describe agent help"
```

Expected: commit succeeds with only test changes. Existing pre-task user edits must not be reverted.

---

### Task 2: Add Public Agent Types And Human Help Flag

**Files:**
- Modify: `yargs.go`

- [ ] **Step 1: Update constants and sentinel errors**

In `yargs.go`, replace the help flag constants and LLM sentinel with:

```go
// Help flag constants
const (
	helpFlagLong  = "--help"
	helpFlagShort = "-h"
	helpFlagAgent = "--help-agent"
	helpCommand   = "help"
)
```

```go
// ErrHelpAgent is returned when agent-readable help is requested (--help-agent).
ErrHelpAgent = errors.New("agent help requested")
```

Keep `ErrHelp`, `ErrSubCommandHelp`, and `ErrShown`.

- [ ] **Step 2: Add AgentInfo and replace LLM metadata**

Replace the three metadata structs with:

```go
// AgentInfo contains optional guidance for agent-readable help output.
type AgentInfo struct {
	Summary   string
	Rules     []string
	Safety    []string
	Discovery []string
}

// CommandInfo contains metadata about the CLI command for help generation.
type CommandInfo struct {
	Name        string
	Description string
	Examples    []string
	Agent       AgentInfo
}

// SubCommandInfo contains metadata about a subcommand for help generation.
type SubCommandInfo struct {
	Name        string
	Description string
	Usage       string // e.g., "SERVICE" or "SERVICE [SERVICE...]"
	Examples    []string
	Aliases     []string
	Hidden      bool // Hidden subcommands don't appear in help but still work
	Agent       AgentInfo
}

// GroupInfo contains metadata about a command group for help generation.
type GroupInfo struct {
	Name        string
	Description string
	Commands    map[string]SubCommandInfo // Commands within this group
	Hidden      bool                      // Hidden groups don't appear in help but still work
	Agent       AgentInfo
}
```

- [ ] **Step 3: Add FlagsSchema to CommandSpec and ResolvedCommand**

Update the registry types:

```go
// CommandSpec describes a subcommand with optional flag and positional schemas.
type CommandSpec struct {
	Info        SubCommandInfo
	FlagsSchema any
	ArgsSchema  any
}
```

Update `ResolvedCommand`:

```go
// FlagsSchema is an optional schema (struct with `flag` tags) used for introspection.
FlagsSchema any
// ArgsSchema is an optional schema (struct with `pos` tags) used for introspection.
ArgsSchema any
```

Update `ResolveCommandWithRegistry` so it attaches both schemas:

```go
if spec, ok := reg.CommandSpec(res.Path); ok {
	res.FlagsSchema = spec.FlagsSchema
	res.ArgsSchema = spec.ArgsSchema
}
```

- [ ] **Step 4: Add help-agent to human help output**

In `GenerateGlobalHelp`, change the help option rows to:

```go
b.WriteString(fmt.Sprintf("%-28s %s\n", fmt.Sprintf("    %s, %s", helpFlagShort, helpFlagLong), "Show help"))
b.WriteString(fmt.Sprintf("%-28s %s\n\n", fmt.Sprintf("        %s", helpFlagAgent), "Show agent-readable CLI context"))
```

In `GenerateSubCommandHelp`, change the same rows to:

```go
b.WriteString(fmt.Sprintf("%-28s %s\n", fmt.Sprintf("    %s, %s", helpFlagShort, helpFlagLong), "Show this help message"))
b.WriteString(fmt.Sprintf("%-28s %s\n\n", fmt.Sprintf("        %s", helpFlagAgent), "Show agent-readable CLI context"))
```

- [ ] **Step 5: Run focused tests**

Run:

```bash
go test ./... -run 'TestGenerateGlobalHelpAdvertisesAgentHelp|TestGenerateAgentHelp' -count=1
```

Expected: `TestGenerateGlobalHelpAdvertisesAgentHelp` passes. Agent generator tests still fail with undefined generator functions.

- [ ] **Step 6: Commit public model changes**

Run:

```bash
git add yargs.go yargs_test.go
git commit -m "help: add agent metadata"
```

Expected: commit succeeds. The test suite can still be red because the generator is not implemented yet.

---

### Task 3: Implement Agent Help Rendering

**Files:**
- Modify: `yargs.go`

- [ ] **Step 1: Add target types and shared defaults**

Add these definitions near the registry types:

```go
type AgentHelpTargetKind string

const (
	AgentHelpGlobal       AgentHelpTargetKind = "global"
	AgentHelpGroup        AgentHelpTargetKind = "group"
	AgentHelpCommand      AgentHelpTargetKind = "command"
	AgentHelpGroupCommand AgentHelpTargetKind = "group_command"
)

type AgentHelpTarget struct {
	Kind        AgentHelpTargetKind
	Path        []string
	Command     CommandInfo
	Group       GroupInfo
	SubCommand  SubCommandInfo
	FlagsSchema any
	ArgsSchema  any
}

var defaultAgentRules = []string{
	"Prefer exact examples when they match the task.",
	"Use command-specific agent help before running an unfamiliar command.",
	"Do not invent flags; use only flags listed in this context or command help.",
	"Preserve arguments after `--` as payload or application arguments.",
}
```

- [ ] **Step 2: Add public generator entrypoints**

Add these functions after `GenerateSubCommandHelpFromConfig`:

```go
func GenerateAgentHelp[G any](config HelpConfig, globalFlagsExample G) string {
	target := AgentHelpTarget{
		Kind:    AgentHelpGlobal,
		Command: config.Command,
	}
	return renderAgentHelp(config, target, globalFlagsExample)
}

func GenerateAgentHelpForCommand[G any, S any, A any](config HelpConfig, commandPath []string, globalFlagsExample G, flagsExample S, argsExample A) string {
	target, ok := agentHelpTargetFromConfig(config, commandPath)
	if !ok {
		return unknownAgentHelpTarget(config.Command.Name, commandPath)
	}
	target.FlagsSchema = flagsExample
	target.ArgsSchema = argsExample
	return renderAgentHelp(config, target, globalFlagsExample)
}

func GenerateAgentHelpFromRegistry[G any](reg Registry, commandPath []string, globalFlagsExample G) string {
	config := reg.HelpConfig()
	target, ok := agentHelpTargetFromConfig(config, commandPath)
	if !ok {
		return unknownAgentHelpTarget(config.Command.Name, commandPath)
	}
	if spec, ok := reg.CommandSpec(target.Path); ok {
		target.FlagsSchema = spec.FlagsSchema
		target.ArgsSchema = spec.ArgsSchema
	}
	return renderAgentHelp(config, target, globalFlagsExample)
}
```

- [ ] **Step 3: Add target resolution helpers**

Add:

```go
func agentHelpTargetFromConfig(config HelpConfig, commandPath []string) (AgentHelpTarget, bool) {
	if len(commandPath) == 0 {
		return AgentHelpTarget{Kind: AgentHelpGlobal, Command: config.Command}, true
	}
	if len(commandPath) == 1 {
		name := canonicalFlatCommandName(config, commandPath[0])
		if cmd, ok := config.SubCommands[name]; ok {
			return AgentHelpTarget{Kind: AgentHelpCommand, Path: []string{name}, Command: config.Command, SubCommand: cmd}, true
		}
		if group, ok := config.Groups[commandPath[0]]; ok {
			return AgentHelpTarget{Kind: AgentHelpGroup, Path: []string{commandPath[0]}, Command: config.Command, Group: group}, true
		}
		return AgentHelpTarget{}, false
	}
	if len(commandPath) == 2 {
		groupName := commandPath[0]
		group, ok := config.Groups[groupName]
		if !ok {
			return AgentHelpTarget{}, false
		}
		cmdName := canonicalGroupCommandName(group, commandPath[1])
		cmd, ok := group.Commands[cmdName]
		if !ok {
			return AgentHelpTarget{}, false
		}
		return AgentHelpTarget{Kind: AgentHelpGroupCommand, Path: []string{groupName, cmdName}, Command: config.Command, Group: group, SubCommand: cmd}, true
	}
	return AgentHelpTarget{}, false
}

func canonicalFlatCommandName(config HelpConfig, name string) string {
	if _, ok := config.SubCommands[name]; ok {
		return name
	}
	for cmdName, info := range config.SubCommands {
		for _, alias := range info.Aliases {
			if alias == name {
				return cmdName
			}
		}
	}
	return name
}

func canonicalGroupCommandName(group GroupInfo, name string) string {
	if _, ok := group.Commands[name]; ok {
		return name
	}
	for cmdName, info := range group.Commands {
		for _, alias := range info.Aliases {
			if alias == name {
				return cmdName
			}
		}
	}
	return name
}

func unknownAgentHelpTarget(commandName string, path []string) string {
	target := strings.Join(path, " ")
	if target == "" {
		target = "<empty>"
	}
	return fmt.Sprintf("Unknown command: %s. Run `%s --help-agent` to inspect available commands.\n", target, commandName)
}
```

- [ ] **Step 4: Add Markdown render helpers**

Add focused helpers below the target resolution helpers:

```go
func renderAgentHelp[G any](config HelpConfig, target AgentHelpTarget, globalFlagsExample G) string {
	var b strings.Builder
	switch target.Kind {
	case AgentHelpGlobal:
		renderGlobalAgentHelp(&b, config, globalFlagsExample)
	case AgentHelpGroup:
		renderGroupAgentHelp(&b, config, target, globalFlagsExample)
	case AgentHelpCommand, AgentHelpGroupCommand:
		renderCommandAgentHelp(&b, config, target, globalFlagsExample)
	}
	return b.String()
}

func renderAgentList(b *strings.Builder, values []string) {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(value)
		b.WriteString("\n")
	}
}

func combinedAgentRules(info AgentInfo) []string {
	rules := make([]string, 0, len(defaultAgentRules)+len(info.Rules))
	rules = append(rules, defaultAgentRules...)
	rules = append(rules, info.Rules...)
	return rules
}

func agentPurpose(description string, info AgentInfo) string {
	if info.Summary != "" {
		return info.Summary
	}
	return description
}
```

- [ ] **Step 5: Implement global rendering**

Add:

```go
func renderGlobalAgentHelp[G any](b *strings.Builder, config HelpConfig, globalFlagsExample G) {
	fmt.Fprintf(b, "# %s Agent Context\n\n", config.Command.Name)
	if purpose := agentPurpose(config.Command.Description, config.Command.Agent); purpose != "" {
		b.WriteString("## Purpose\n\n")
		b.WriteString(purpose)
		b.WriteString("\n\n")
	}
	b.WriteString("## Operating Rules\n\n")
	renderAgentList(b, combinedAgentRules(config.Command.Agent))
	b.WriteString("\n")
	if len(config.Command.Agent.Discovery) > 0 || len(config.SubCommands) > 0 || len(config.Groups) > 0 {
		b.WriteString("## Discovery\n\n")
		renderAgentList(b, config.Command.Agent.Discovery)
		for _, name := range visibleSubcommandNames(config.SubCommands) {
			fmt.Fprintf(b, "- Run `%s %s --help-agent` for command-specific context.\n", config.Command.Name, name)
		}
		for _, name := range visibleGroupNames(config.Groups) {
			fmt.Fprintf(b, "- Run `%s %s --help-agent` for group-specific context.\n", config.Command.Name, name)
		}
		b.WriteString("\n")
	}
	renderAgentFlagsSection(b, "Global Options", extractFlagInfo(reflect.TypeOf(globalFlagsExample)))
	renderAgentCommandsSection(b, config)
	renderAgentGroupsSection(b, config)
	renderAgentExamplesSection(b, config.Command.Examples)
	renderAgentSafetySection(b, config.Command.Agent.Safety)
}
```

Add these visible-name helpers:

```go
func visibleSubcommandNames(commands map[string]SubCommandInfo) []string {
	names := make([]string, 0, len(commands))
	for name, info := range commands {
		if !info.Hidden {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

func visibleGroupNames(groups map[string]GroupInfo) []string {
	names := make([]string, 0, len(groups))
	for name, info := range groups {
		if !info.Hidden {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}
```

- [ ] **Step 6: Implement group and command rendering**

Add group rendering that uses the same helper sections:

```go
func renderGroupAgentHelp[G any](b *strings.Builder, config HelpConfig, target AgentHelpTarget, globalFlagsExample G) {
	groupName := target.Path[0]
	fmt.Fprintf(b, "# %s %s Agent Context\n\n", config.Command.Name, groupName)
	if purpose := agentPurpose(target.Group.Description, target.Group.Agent); purpose != "" {
		b.WriteString("## Purpose\n\n")
		b.WriteString(purpose)
		b.WriteString("\n\n")
	}
	b.WriteString("## Operating Rules\n\n")
	renderAgentList(b, combinedAgentRules(target.Group.Agent))
	b.WriteString("\n")
	b.WriteString("## Discovery\n\n")
	renderAgentList(b, target.Group.Agent.Discovery)
	for _, name := range visibleSubcommandNames(target.Group.Commands) {
		fmt.Fprintf(b, "- Run `%s %s %s --help-agent` for command-specific context.\n", config.Command.Name, groupName, name)
	}
	b.WriteString("\n")
	renderAgentGroupCommandsSection(b, groupName, target.Group)
	renderAgentFlagsSection(b, "Global Options", extractFlagInfo(reflect.TypeOf(globalFlagsExample)))
	renderAgentSafetySection(b, target.Group.Agent.Safety)
}
```

Add command rendering:

```go
func renderCommandAgentHelp[G any](b *strings.Builder, config HelpConfig, target AgentHelpTarget, globalFlagsExample G) {
	path := strings.Join(target.Path, " ")
	fmt.Fprintf(b, "# %s %s Agent Context\n\n", config.Command.Name, path)
	if purpose := agentPurpose(target.SubCommand.Description, target.SubCommand.Agent); purpose != "" {
		b.WriteString("## Purpose\n\n")
		b.WriteString(purpose)
		b.WriteString("\n\n")
	}
	if len(target.SubCommand.Aliases) > 0 {
		b.WriteString("**Aliases**: ")
		for i, alias := range target.SubCommand.Aliases {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, "`%s`", alias)
		}
		b.WriteString("\n\n")
	}
	b.WriteString("## Usage\n\n")
	b.WriteString("```\n")
	b.WriteString(agentUsage(config.Command.Name, target))
	b.WriteString("\n```\n\n")
	renderAgentArgsSection(b, extractArgsInfo(reflect.TypeOf(target.ArgsSchema)))
	renderAgentFlagsSection(b, "Options", extractFlagInfo(reflect.TypeOf(target.FlagsSchema)))
	renderAgentFlagsSection(b, "Global Options", extractFlagInfo(reflect.TypeOf(globalFlagsExample)))
	renderAgentExamplesSection(b, target.SubCommand.Examples)
	renderAgentSafetySection(b, target.SubCommand.Agent.Safety)
}
```

Add `agentUsage`:

```go
func agentUsage(commandName string, target AgentHelpTarget) string {
	if target.Kind == AgentHelpGroupCommand {
		groupName := target.Path[0]
		cmdName := target.Path[1]
		if target.SubCommand.Usage != "" {
			return formatGroupCommandUsage(commandName, groupName, cmdName, target.SubCommand.Usage)
		}
		return fmt.Sprintf("%s [GLOBAL_OPTIONS] %s %s [ARGS...]", commandName, groupName, cmdName)
	}

	cmdName := target.Path[0]
	argsInfo := extractArgsInfo(reflect.TypeOf(target.ArgsSchema))
	usage := fmt.Sprintf("%s [GLOBAL_OPTIONS] %s", commandName, cmdName)
	for _, arg := range argsInfo {
		argName := strings.ToUpper(arg.Name)
		if arg.Variadic {
			if arg.MinCount > 0 {
				usage += fmt.Sprintf(" <%s...>", argName)
			} else {
				usage += fmt.Sprintf(" [%s...]", argName)
			}
			continue
		}
		if arg.Required {
			usage += fmt.Sprintf(" <%s>", argName)
		} else {
			usage += fmt.Sprintf(" [%s]", argName)
		}
	}
	usage += " [OPTIONS]"
	if target.SubCommand.Usage != "" {
		usage += " " + target.SubCommand.Usage
	}
	return usage
}
```

- [ ] **Step 7: Add section helpers for args, flags, commands, examples, and safety**

Add these helpers:

```go
func renderAgentArgsSection(b *strings.Builder, args []argInfo) {
	if len(args) == 0 {
		return
	}
	b.WriteString("## Arguments\n\n")
	for _, arg := range args {
		argName := strings.ToUpper(arg.Name)
		fmt.Fprintf(b, "### `%s`\n\n", argName)
		if arg.Description != "" {
			b.WriteString(arg.Description)
			b.WriteString("\n\n")
		}
		fmt.Fprintf(b, "- **Type**: `%s`\n", arg.Type)
		fmt.Fprintf(b, "- **Required**: %v\n", arg.Required)
		if arg.Variadic {
			fmt.Fprintf(b, "- **Variadic**: true (minimum: %d)\n", arg.MinCount)
		}
		b.WriteString("\n")
	}
}
```

```go
func renderAgentFlagsSection(b *strings.Builder, title string, flags []flagInfo) {
	if len(flags) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", title)
	for _, flag := range flags {
		fmt.Fprintf(b, "### `--%s`", flag.Name)
		if flag.ShortName != "" {
			fmt.Fprintf(b, " (short: `-%s`)", flag.ShortName)
		}
		b.WriteString("\n\n")
		if flag.Description != "" {
			b.WriteString(flag.Description)
			b.WriteString("\n\n")
		}
		fmt.Fprintf(b, "- **Type**: `%s`\n", flag.Type)
		if flag.DefaultVal != "" {
			fmt.Fprintf(b, "- **Default**: `%s`\n", flag.DefaultVal)
		}
		b.WriteString("\n")
	}
}
```

```go
func renderAgentCommandsSection(b *strings.Builder, config HelpConfig) {
	names := visibleSubcommandNames(config.SubCommands)
	if len(names) == 0 {
		return
	}
	b.WriteString("## Commands\n\n")
	for _, name := range names {
		info := config.SubCommands[name]
		fmt.Fprintf(b, "### `%s`\n\n", name)
		if info.Description != "" {
			b.WriteString(info.Description)
			b.WriteString("\n\n")
		}
		if len(info.Aliases) > 0 {
			b.WriteString("**Aliases**: ")
			for i, alias := range info.Aliases {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(b, "`%s`", alias)
			}
			b.WriteString("\n\n")
		}
		if purpose := agentPurpose("", info.Agent); purpose != "" {
			b.WriteString(purpose)
			b.WriteString("\n\n")
		}
		fmt.Fprintf(b, "Run `%s %s --help-agent` for command-specific context.\n\n", config.Command.Name, name)
	}
}

func renderAgentGroupsSection(b *strings.Builder, config HelpConfig) {
	names := visibleGroupNames(config.Groups)
	if len(names) == 0 {
		return
	}
	b.WriteString("## Command Groups\n\n")
	for _, name := range names {
		info := config.Groups[name]
		fmt.Fprintf(b, "### `%s`\n\n", name)
		if info.Description != "" {
			b.WriteString(info.Description)
			b.WriteString("\n\n")
		}
		if purpose := agentPurpose("", info.Agent); purpose != "" {
			b.WriteString(purpose)
			b.WriteString("\n\n")
		}
		fmt.Fprintf(b, "Run `%s %s --help-agent` for group-specific context.\n\n", config.Command.Name, name)
	}
}
```

```go
func renderAgentGroupCommandsSection(b *strings.Builder, groupName string, group GroupInfo) {
	names := visibleSubcommandNames(group.Commands)
	if len(names) == 0 {
		return
	}
	b.WriteString("## Commands\n\n")
	for _, name := range names {
		info := group.Commands[name]
		fmt.Fprintf(b, "### `%s %s`\n\n", groupName, name)
		if info.Description != "" {
			b.WriteString(info.Description)
			b.WriteString("\n\n")
		}
		if len(info.Aliases) > 0 {
			b.WriteString("**Aliases**: ")
			for i, alias := range info.Aliases {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(b, "`%s`", alias)
			}
			b.WriteString("\n\n")
		}
	}
}

func renderAgentExamplesSection(b *strings.Builder, examples []string) {
	if len(examples) == 0 {
		return
	}
	b.WriteString("## Examples\n\n")
	for _, example := range examples {
		fmt.Fprintf(b, "```\n%s\n```\n\n", example)
	}
}

func renderAgentSafetySection(b *strings.Builder, safety []string) {
	if len(safety) == 0 {
		return
	}
	b.WriteString("## Safety Notes\n\n")
	renderAgentList(b, safety)
	b.WriteString("\n")
}
```

- [ ] **Step 8: Run focused generator tests**

Run:

```bash
go test ./... -run 'TestGenerateAgentHelp' -count=1
```

Expected: agent generation tests pass. Dispatch and old LLM removal tests may still fail.

- [ ] **Step 9: Commit generator implementation**

Run:

```bash
git add yargs.go yargs_agent_test.go
git commit -m "help: generate agent context"
```

Expected: commit succeeds.

---

### Task 4: Wire Agent Help Into Parsing And Dispatch

**Files:**
- Modify: `yargs.go`
- Modify: `yargs_agent_test.go`
- Modify: `yargs_test.go`

- [ ] **Step 1: Add failing parse-and-dispatch tests**

Append these tests to `yargs_agent_test.go`:

```go
func TestErrHelpAgent(t *testing.T) {
	type GlobalFlags struct{}
	type RunFlags struct{}
	type RunArgs struct{}

	config := HelpConfig{
		Command: CommandInfo{Name: "testcli"},
		SubCommands: map[string]SubCommandInfo{
			"run": {Name: "run", Description: "Run command"},
		},
	}

	tests := []struct {
		name string
		args []string
	}{
		{name: "global help-agent", args: []string{"--help-agent"}},
		{name: "subcommand help-agent", args: []string{"run", "--help-agent"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseWithCommandAndHelp[GlobalFlags, RunFlags, RunArgs](tt.args, config)
			if err != ErrHelpAgent {
				t.Fatalf("expected ErrHelpAgent, got %v", err)
			}
			if result == nil || result.HelpText == "" {
				t.Fatalf("expected non-empty agent help result")
			}
			if strings.Contains(result.HelpText, "--help-llm") {
				t.Fatalf("agent help should not mention --help-llm:\n%s", result.HelpText)
			}
		})
	}
}

func TestParseAndHandleHelpAgentReturnsErrShown(t *testing.T) {
	type GlobalFlags struct{}
	type RunFlags struct{}
	type RunArgs struct{}

	config := HelpConfig{
		Command: CommandInfo{Name: "testcli"},
		SubCommands: map[string]SubCommandInfo{
			"run": {Name: "run", Description: "Run command"},
		},
	}

	result, err := ParseAndHandleHelp[GlobalFlags, RunFlags, RunArgs]([]string{"--help-agent"}, config)
	if err != ErrShown {
		t.Fatalf("expected ErrShown after handling agent help, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result when ErrShown is returned")
	}
}
```

Add dispatch coverage in `yargs_test.go` near `RunSubcommandsWithGroups` tests:

```go
func TestRunSubcommandsWithGroupsHandlesAgentHelp(t *testing.T) {
	type GlobalFlags struct {
		Verbose bool `flag:"verbose" short:"v" help:"Verbose mode"`
	}

	config := HelpConfig{
		Command: CommandInfo{Name: "myapp", Description: "My app"},
		SubCommands: map[string]SubCommandInfo{
			"status": {Name: "status", Description: "Show status"},
		},
		Groups: map[string]GroupInfo{
			"docker": {
				Name:        "docker",
				Description: "Docker commands",
				Commands: map[string]SubCommandInfo{
					"ps": {Name: "ps", Description: "List containers"},
				},
			},
		},
	}
	commands := map[string]SubcommandHandler{
		"status": func(context.Context, []string) error {
			t.Fatal("status handler should not run for agent help")
			return nil
		},
	}
	groups := map[string]Group{
		"docker": {
			Description: "Docker commands",
			Commands: map[string]SubcommandHandler{
				"ps": func(context.Context, []string) error {
					t.Fatal("docker ps handler should not run for agent help")
					return nil
				},
			},
		},
	}

	for _, args := range [][]string{
		{"--help-agent"},
		{"status", "--help-agent"},
		{"docker", "--help-agent"},
		{"docker", "ps", "--help-agent"},
	} {
		if err := RunSubcommandsWithGroups(context.Background(), args, config, GlobalFlags{}, commands, groups); err != nil {
			t.Fatalf("RunSubcommandsWithGroups(%v) returned error: %v", args, err)
		}
	}
}
```

If `context` is not already imported in `yargs_test.go`, add it to that file's import list.

- [ ] **Step 2: Verify tests fail before implementation**

Run:

```bash
go test ./... -run 'TestErrHelpAgent|TestParseAndHandleHelpAgent|TestRunSubcommandsWithGroupsHandlesAgentHelp' -count=1
```

Expected: FAIL because parse and dispatch still check the old agent flag path or do not recognize `ErrHelpAgent`.

- [ ] **Step 3: Update help flag helpers**

Replace `isHelpFlag`, `helpFlagsInArgs`, and `groupedHelpTargets` with agent-aware names:

```go
func isHelpFlag(arg string) bool {
	return arg == helpCommand || arg == helpFlagLong || arg == helpFlagShort
}

func helpFlagsInArgs(args []string) (help bool, agentHelp bool) {
	for _, arg := range args {
		if arg == helpFlagAgent {
			agentHelp = true
			continue
		}
		if arg == helpFlagLong || arg == helpFlagShort {
			help = true
		}
	}
	return help, agentHelp
}

func groupedHelpTargets(args []string) (groupHelp bool, groupAgentHelp bool, commandHelp bool, commandAgentHelp bool) {
	indices := findNonFlagArgs(args)
	if len(indices) < 2 {
		groupHelp, groupAgentHelp = helpFlagsInArgs(args)
		return groupHelp, groupAgentHelp, false, false
	}
	groupHelp, groupAgentHelp = helpFlagsInArgs(args[indices[0]+1 : indices[1]])
	commandHelp, commandAgentHelp = helpFlagsInArgs(args[indices[1]+1:])
	return groupHelp, groupAgentHelp, commandHelp, commandAgentHelp
}
```

Update `ResolveCommand` so the early help check uses `helpFlagAgent`.

- [ ] **Step 4: Update ParseWithCommandAndHelp**

Change the global agent help branch:

```go
if args[0] == helpFlagAgent {
	helpText := GenerateAgentHelp(config, globalFlags)
	return &TypedParseResult[G, S, A]{HelpText: helpText}, ErrHelpAgent
}
```

Change subcommand agent help handling:

```go
if args[i] == helpFlagAgent {
	helpText := GenerateAgentHelpForCommand(config, []string{subCmdName}, globalFlags, subCmdFlags, argsStruct)
	return &TypedParseResult[G, S, A]{HelpText: helpText}, ErrHelpAgent
}
```

Update comments to refer to `--help-agent`.

- [ ] **Step 5: Update RunSubcommandsWithGroups**

Replace global agent branch with:

```go
if args[0] == helpFlagAgent {
	fmt.Print(GenerateAgentHelp(config, globalFlagsExample))
	return nil
}
```

In grouped dispatch, replace group agent help generation with:

```go
fmt.Print(GenerateAgentHelpForCommand(config, []string{first}, globalFlagsExample, struct{}{}, struct{}{}))
```

Replace grouped command agent help generation with:

```go
fmt.Print(GenerateAgentHelpForCommand(config, []string{first, second}, globalFlagsExample, struct{}{}, struct{}{}))
```

Replace flat command agent help generation with:

```go
fmt.Print(GenerateAgentHelpForCommand(config, []string{cmdName}, globalFlagsExample, struct{}{}, struct{}{}))
```

- [ ] **Step 6: Update ParseAndHandleHelp**

In the help sentinel handling branch, replace `ErrHelpLLM` with `ErrHelpAgent`:

```go
if errors.Is(err, ErrHelp) || errors.Is(err, ErrSubCommandHelp) || errors.Is(err, ErrHelpAgent) {
	fmt.Print(result.HelpText)
	return nil, ErrShown
}
```

- [ ] **Step 7: Run focused dispatch tests**

Run:

```bash
go test ./... -run 'TestErrHelpAgent|TestParseAndHandleHelpAgent|TestRunSubcommandsWithGroupsHandlesAgentHelp' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit dispatch integration**

Run:

```bash
git add yargs.go yargs_agent_test.go yargs_test.go
git commit -m "help: route agent help"
```

Expected: commit succeeds.

---

### Task 5: Remove Old LLM Surface Completely

**Files:**
- Modify: `yargs.go`
- Modify: `yargs_agent_test.go`
- Modify: `yargs_test.go`

- [ ] **Step 1: Add removal assertion**

Append this test to `yargs_agent_test.go`:

```go
func TestHelpLLMIsRemoved(t *testing.T) {
	type GlobalFlags struct{}
	type RunFlags struct{}
	type RunArgs struct{}

	config := HelpConfig{
		Command: CommandInfo{Name: "testcli"},
		SubCommands: map[string]SubCommandInfo{
			"run": {Name: "run", Description: "Run command"},
		},
	}

	result, err := ParseWithCommandAndHelp[GlobalFlags, RunFlags, RunArgs]([]string{"--help-llm"}, config)
	if err == nil {
		t.Fatalf("expected --help-llm to be rejected")
	}
	if result != nil && result.HelpText != "" {
		t.Fatalf("removed --help-llm should not produce help text:\n%s", result.HelpText)
	}
}
```

- [ ] **Step 2: Remove old LLM generators and references**

Delete these functions from `yargs.go`:

- `GenerateGlobalHelpLLM`
- `GenerateGroupHelpLLM`
- `GenerateGroupCommandHelpLLM`
- `GenerateSubCommandHelpLLM`
- `GenerateSubCommandHelpLLMFromConfig`

Run:

```bash
rg -n "HelpLLM|help-llm|LLMInstructions|ErrHelpLLM|helpFlagLLM|llm" yargs.go yargs_agent_test.go yargs_test.go
```

Expected: no matches except intentional lowercase words in comments outside the removed surface. If matches remain in API names, remove them.

- [ ] **Step 3: Run full test suite**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Commit removal**

Run:

```bash
git add yargs.go yargs_agent_test.go yargs_test.go
git commit -m "help: remove llm help surface"
```

Expected: commit succeeds.

---

### Task 6: Update Documentation

**Files:**
- Modify: `README.md`
- Modify: `doc.go`

- [ ] **Step 1: Update README help section**

Replace the `Help Generation (Human + LLM)` section in `README.md` with:

```markdown
## Help Generation (Human + Agent)

Yargs can emit human help or agent-readable help from the same metadata.
Agent help is Markdown intended to be loaded directly into Codex-style agents
that are operating a CLI through normal command execution.

### Human help

- Global: `GenerateGlobalHelp`
- Group: `GenerateGroupHelp`
- Group command: `GenerateGroupCommandHelp`
- Subcommand: `GenerateSubCommandHelp`
- Dispatcher: `RunSubcommands` and `RunSubcommandsWithGroups`

### Agent help

- Global: `GenerateAgentHelp`
- Command path: `GenerateAgentHelpForCommand`
- Registry-backed command path: `GenerateAgentHelpFromRegistry`
- Flag: `--help-agent`

Use `AgentInfo` when a CLI needs to provide extra operating rules, safety notes,
or discovery guidance for agents. Yargs still generates useful agent help from
descriptions, examples, aliases, flags, defaults, and positional argument tags
when `AgentInfo` is empty.

### Parse-and-handle help

`ParseWithCommandAndHelp` and `ParseAndHandleHelp` will detect `help`, `-h`,
`--help`, and `--help-agent` and return the right error sentinel.
`ParseAndHandleHelp` prints help automatically and returns `ErrShown`.
The `help` subcommand is supported as `app help` or `app help <command>`.

### Help metadata fields

You control help output with these fields:

- `CommandInfo`: `Name`, `Description`, `Examples`, `Agent`
- `SubCommandInfo`: `Name`, `Description`, `Usage`, `Examples`, `Aliases`, `Hidden`, `Agent`
- `GroupInfo`: `Name`, `Description`, `Commands`, `Hidden`, `Agent`
- `AgentInfo`: `Summary`, `Rules`, `Safety`, `Discovery`
```

- [ ] **Step 2: Update README registry example**

In the registry example, add `FlagsSchema`:

```go
type RunFlags struct {
    Detach bool `flag:"detach" short:"d" help:"Run in background"`
}

reg := yargs.Registry{
    Command: yargs.CommandInfo{Name: "app"},
    SubCommands: map[string]yargs.CommandSpec{
        "run": {
            Info:        yargs.SubCommandInfo{Name: "run"},
            FlagsSchema: RunFlags{},
            ArgsSchema:  RunArgs{},
        },
    },
}
```

- [ ] **Step 3: Update package docs**

In `doc.go`, update package comments so they mention agent help:

```go
//   - Automatic help generation for humans and agents
```

Add an agent metadata example in the help config snippet:

```go
Command: yargs.CommandInfo{
    Name:        "myapp",
    Description: "My application",
    Agent: yargs.AgentInfo{
        Summary: "Use myapp to inspect and operate services.",
        Rules: []string{"Use command-specific agent help before unfamiliar commands."},
    },
},
```

In the registry docs, mention that `CommandSpec.FlagsSchema` and
`CommandSpec.ArgsSchema` let agent help include detailed options and positional
arguments.

- [ ] **Step 4: Verify docs no longer mention removed LLM API**

Run:

```bash
rg -n "help-llm|HelpLLM|LLMInstructions|ErrHelpLLM|LLM help|LLM-optimized" README.md doc.go yargs.go yargs_agent_test.go yargs_test.go
```

Expected: no matches.

- [ ] **Step 5: Run checks**

Run:

```bash
mise run check
```

Expected: PASS.

- [ ] **Step 6: Commit docs**

Run:

```bash
git add README.md doc.go
git commit -m "docs: document agent help"
```

Expected: commit succeeds.

---

### Task 7: Final Verification

**Files:**
- Verify: all changed files

- [ ] **Step 1: Inspect diff for removed API and new API**

Run:

```bash
rg -n "help-agent|AgentInfo|ErrHelpAgent|GenerateAgentHelp|FlagsSchema" yargs.go yargs_agent_test.go README.md doc.go
```

Expected: matches show the new public surface in code, tests, and docs.

- [ ] **Step 2: Confirm old API is gone**

Run:

```bash
rg -n "help-llm|HelpLLM|LLMInstructions|ErrHelpLLM|helpFlagLLM|LLM-optimized" .
```

Expected: no matches.

- [ ] **Step 3: Run the full project check**

Run:

```bash
mise run check
```

Expected: PASS.

- [ ] **Step 4: Inspect git status**

Run:

```bash
git status --short
```

Expected: clean, or only files intentionally left unstaged by the user. If user-owned pre-existing edits remain, report them separately and do not revert them.

- [ ] **Step 5: Final commit for verification changes**

Check for verification changes:

```bash
git status --short
```

When formatting or tidy changed tracked files after the docs commit, run:

```bash
git add yargs.go yargs_agent_test.go yargs_test.go README.md doc.go go.mod
git commit -m "chore: finalize agent help"
```

Expected: commit succeeds when there were final verification changes. When `git status --short` is clean before this step, record that no final commit was needed.
