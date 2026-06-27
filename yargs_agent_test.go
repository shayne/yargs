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

func TestGenerateAgentHelpHiddenTargetsAreUnknown(t *testing.T) {
	config := HelpConfig{
		Command: CommandInfo{Name: "testcli"},
		SubCommands: map[string]SubCommandInfo{
			"secret": {
				Name:        "secret",
				Description: "Hidden flat description",
				Aliases:     []string{"sec"},
				Hidden:      true,
			},
		},
		Groups: map[string]GroupInfo{
			"admin": {
				Name:        "admin",
				Description: "Hidden group description",
				Hidden:      true,
				Commands: map[string]SubCommandInfo{
					"wipe": {
						Name:        "wipe",
						Description: "Hidden group child description",
					},
				},
			},
			"docker": {
				Name:        "docker",
				Description: "Docker commands",
				Commands: map[string]SubCommandInfo{
					"hidden": {
						Name:        "hidden",
						Description: "Hidden grouped command description",
						Aliases:     []string{"hid"},
						Hidden:      true,
					},
				},
			},
		},
	}

	tests := []struct {
		name string
		path []string
		desc string
	}{
		{name: "hidden flat", path: []string{"secret"}, desc: "Hidden flat description"},
		{name: "hidden flat alias", path: []string{"sec"}, desc: "Hidden flat description"},
		{name: "hidden group", path: []string{"admin"}, desc: "Hidden group description"},
		{name: "child inside hidden group", path: []string{"admin", "wipe"}, desc: "Hidden group child description"},
		{name: "hidden grouped child", path: []string{"docker", "hidden"}, desc: "Hidden grouped command description"},
		{name: "hidden grouped child alias", path: []string{"docker", "hid"}, desc: "Hidden grouped command description"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := GenerateAgentHelpForCommand(config, tt.path, struct{}{}, struct{}{}, struct{}{})
			mustContain(t, output, "Unknown command: "+strings.Join(tt.path, " ")+".", "Run `testcli --help-agent` to inspect available commands.")
			mustNotContain(t, output, tt.desc)
		})
	}

	reg := Registry{
		Command: config.Command,
		SubCommands: map[string]CommandSpec{
			"secret": {Info: config.SubCommands["secret"]},
		},
		Groups: map[string]GroupSpec{
			"admin": {
				Info: GroupInfo{
					Name:        "admin",
					Description: "Hidden registry group description",
					Hidden:      true,
				},
				Commands: map[string]CommandSpec{
					"wipe": {Info: SubCommandInfo{Name: "wipe", Description: "Hidden registry child description"}},
				},
			},
		},
	}
	output := GenerateAgentHelpFromRegistry(reg, []string{"admin", "wipe"}, struct{}{})
	mustContain(t, output, "Unknown command: admin wipe.")
	mustNotContain(t, output, "Hidden registry child description")
}

func TestGenerateAgentHelpIncludesInheritedSafety(t *testing.T) {
	config := HelpConfig{
		Command: CommandInfo{
			Name: "testcli",
			Agent: AgentInfo{
				Safety: []string{"Global safety guidance."},
			},
		},
		SubCommands: map[string]SubCommandInfo{
			"remove": {
				Name: "remove",
				Agent: AgentInfo{
					Safety: []string{"Command safety guidance."},
				},
			},
		},
		Groups: map[string]GroupInfo{
			"docker": {
				Name: "docker",
				Agent: AgentInfo{
					Safety: []string{"Group safety guidance."},
				},
				Commands: map[string]SubCommandInfo{
					"prune": {
						Name: "prune",
						Agent: AgentInfo{
							Safety: []string{"Grouped command safety guidance."},
						},
					},
				},
			},
		},
	}

	commandOutput := GenerateAgentHelpForCommand(config, []string{"remove"}, struct{}{}, struct{}{}, struct{}{})
	mustContain(t, commandOutput,
		"## Safety Notes",
		"Global safety guidance.",
		"Command safety guidance.",
	)

	groupOutput := GenerateAgentHelpForCommand(config, []string{"docker"}, struct{}{}, struct{}{}, struct{}{})
	mustContain(t, groupOutput,
		"## Safety Notes",
		"Global safety guidance.",
		"Group safety guidance.",
	)

	groupCommandOutput := GenerateAgentHelpForCommand(config, []string{"docker", "prune"}, struct{}{}, struct{}{}, struct{}{})
	mustContain(t, groupCommandOutput,
		"## Safety Notes",
		"Global safety guidance.",
		"Group safety guidance.",
		"Grouped command safety guidance.",
	)
}

func TestGenerateAgentHelpExplicitUsageDoesNotDuplicateArgs(t *testing.T) {
	type RunArgs struct {
		Service string `pos:"0" help:"Service name"`
	}
	config := HelpConfig{
		Command: CommandInfo{Name: "testcli"},
		SubCommands: map[string]SubCommandInfo{
			"run": {
				Name:  "run",
				Usage: "SERVICE [--detach]",
			},
		},
	}

	output := GenerateAgentHelpForCommand(config, []string{"run"}, struct{}{}, struct{}{}, RunArgs{})
	mustContain(t, output, "testcli [GLOBAL_OPTIONS] run SERVICE [--detach]")
	mustNotContain(t, output, "<SERVICE> [OPTIONS] SERVICE")
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
