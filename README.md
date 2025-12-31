# yargs

A reflection-based, generic, type-safe CLI parser for Go that reads struct tags
and generates help text, with subcommands, grouped commands, and LLM-friendly
output.

## Table of Contents

- [Features](#features)
- [Install](#install)
- [Quick Start (Single Command)](#quick-start-single-command)
- [Subcommands (Global + Command Flags)](#subcommands-global--command-flags)
- [Command Groups ("gh repo" style)](#command-groups-gh-repo-style)
- [Help Generation (Human + LLM)](#help-generation-human--llm)
- [Help metadata fields](#help-metadata-fields)
- [Flag Tags](#flag-tags)
- [Supported Flag Types](#supported-flag-types)
- [Optional flags via pointers](#optional-flags-via-pointers)
- [Port validation](#port-validation)
- [Positional Argument Schemas](#positional-argument-schemas)
- [Aliases](#aliases)
- [Partial Parsing (Known Flags)](#partial-parsing-known-flags)
- [Registry & Introspection](#registry--introspection)
- [Remaining Args (`--`)](#remaining-args---)
- [Errors](#errors)
- [API Reference](#api-reference)
- [Development](#development)
- [License](#license)

## Features

- Type-safe flag parsing with Go generics.
- Global + subcommand flag separation (including mixed ordering).
- Positional argument schemas with validation and auto-population.
- Automatic help text (human and LLM optimized).
- Command groups ("docker run"-style) and aliases.
- Flags can appear anywhere; supports `--` passthrough.
- Optional flags via pointer types; defaults via struct tags.
- Built-in `Port` type with optional range validation.
- Partial parsing for "known flags" and consume-only workflows.
- Registry-based command introspection for advanced dispatching.

## Install

```bash
go get github.com/shayne/yargs@latest
```

## Quick Start (Single Command)

Use `ParseFlags` when you have no subcommands.

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/shayne/yargs"
)

type Flags struct {
    Verbose bool   `flag:"verbose" short:"v" help:"Enable verbose output"`
    Output  string `flag:"output" short:"o" help:"Output file"`
}

func main() {
    result, err := yargs.ParseFlags[Flags](os.Args[1:])
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("verbose=%v output=%s args=%v\n", result.Flags.Verbose, result.Flags.Output, result.Args)
}
```

## Subcommands (Global + Command Flags)

Use `ParseWithCommand` or `ParseAndHandleHelp` when you have subcommands.

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "os"

    "github.com/shayne/yargs"
)

type GlobalFlags struct {
    Verbose bool `flag:"verbose" short:"v" help:"Enable verbose output"`
}

type StatusFlags struct {
    Short bool `flag:"short" short:"s" help:"Show short status"`
}

type CommitFlags struct {
    Message string `flag:"message" short:"m" help:"Commit message"`
    Amend   bool   `flag:"amend" help:"Amend the last commit"`
}

type CloneFlags struct {
    Depth  int    `flag:"depth" help:"Create a shallow clone with history truncated"`
    Branch string `flag:"branch" short:"b" help:"Checkout a specific branch"`
}

type CloneArgs struct {
    Repo string `pos:"0" help:"Repository URL"`
    Dir  string `pos:"1?" help:"Target directory"`
}

var helpConfig = yargs.HelpConfig{
    Command: yargs.CommandInfo{
        Name:        "git",
        Description: "The stupid content tracker",
    },
    SubCommands: map[string]yargs.SubCommandInfo{
        "status": {Name: "status", Description: "Show working tree status"},
        "commit": {Name: "commit", Description: "Record changes to the repository"},
        "clone":  {Name: "clone", Description: "Clone a repository into a new directory"},
    },
}

func handleStatus(args []string) error {
    return runWithParse[StatusFlags, struct{}](args, func(result *yargs.TypedParseResult[GlobalFlags, StatusFlags, struct{}]) error {
        fmt.Printf("short=%v verbose=%v\n", result.SubCommandFlags.Short, result.GlobalFlags.Verbose)
        return nil
    })
}

func handleCommit(args []string) error {
    return runWithParse[CommitFlags, struct{}](args, func(result *yargs.TypedParseResult[GlobalFlags, CommitFlags, struct{}]) error {
        fmt.Printf("message=%q amend=%v\n", result.SubCommandFlags.Message, result.SubCommandFlags.Amend)
        return nil
    })
}

func handleClone(args []string) error {
    return runWithParse[CloneFlags, CloneArgs](args, func(result *yargs.TypedParseResult[GlobalFlags, CloneFlags, CloneArgs]) error {
        fmt.Printf("repo=%s dir=%s depth=%d branch=%s\n",
            result.Args.Repo,
            result.Args.Dir,
            result.SubCommandFlags.Depth,
            result.SubCommandFlags.Branch,
        )
        return nil
    })
}

func runWithParse[S any, A any](args []string, fn func(*yargs.TypedParseResult[GlobalFlags, S, A]) error) error {
    result, err := yargs.ParseAndHandleHelp[GlobalFlags, S, A](args, helpConfig)
    if errors.Is(err, yargs.ErrShown) {
        return nil
    }
    if err != nil {
        return err
    }
    return fn(result)
}

func main() {
    handlers := map[string]yargs.SubcommandHandler{
        "status": func(_ context.Context, args []string) error { return handleStatus(args) },
        "commit": func(_ context.Context, args []string) error { return handleCommit(args) },
        "clone":  func(_ context.Context, args []string) error { return handleClone(args) },
    }
    if err := yargs.RunSubcommands(context.Background(), os.Args[1:], helpConfig, GlobalFlags{}, handlers); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Command Groups ("gh repo" style)

Use `RunSubcommandsWithGroups` and `Group`/`GroupInfo` for grouped commands.

```go
config := yargs.HelpConfig{
    Command: yargs.CommandInfo{Name: "gh", Description: "GitHub CLI"},
    Groups: map[string]yargs.GroupInfo{
        "repo": {
            Name:        "repo",
            Description: "Manage repositories",
            Commands: map[string]yargs.SubCommandInfo{
                "create": {Name: "create", Description: "Create a repository"},
                "view":   {Name: "view", Description: "View a repository"},
            },
        },
        "issue": {
            Name:        "issue",
            Description: "Manage issues",
            Commands: map[string]yargs.SubCommandInfo{
                "list":   {Name: "list", Description: "List issues"},
                "create": {Name: "create", Description: "Create an issue"},
            },
        },
    },
}

groups := map[string]yargs.Group{
    "repo": {
        Description: "Manage repositories",
        Commands: map[string]yargs.SubcommandHandler{
            "create": handleRepoCreate,
            "view":   handleRepoView,
        },
    },
    "issue": {
        Description: "Manage issues",
        Commands: map[string]yargs.SubcommandHandler{
            "list":   handleIssueList,
            "create": handleIssueCreate,
        },
    },
}

_ = yargs.RunSubcommandsWithGroups(context.Background(), os.Args[1:], config, GlobalFlags{}, nil, groups)
```

## Help Generation (Human + LLM)

Yargs can emit human help or LLM-optimized help from the same metadata.

### Human help

- Global: `GenerateGlobalHelp`
- Group: `GenerateGroupHelp`
- Subcommand: `GenerateSubCommandHelp`
- Dispatcher: `RunSubcommands` and `RunSubcommandsWithGroups`

### LLM help

- Global: `GenerateGlobalHelpLLM`
- Group: `GenerateGroupHelpLLM`
- Subcommand: `GenerateSubCommandHelpLLM`
- Flags: `--help-llm`

### Parse-and-handle help

`ParseWithCommandAndHelp` and `ParseAndHandleHelp` will detect `help`, `-h`,
`--help`, and `--help-llm` and return the right error sentinel.
`ParseAndHandleHelp` prints help automatically and returns `ErrShown`.
The `help` subcommand is supported as `app help` or `app help <command>`.

### Help metadata fields

You control help output with these fields:

- `CommandInfo`: `Name`, `Description`, `Examples`, `LLMInstructions`
- `SubCommandInfo`: `Name`, `Description`, `Usage`, `Examples`, `Aliases`, `Hidden`, `LLMInstructions`
- `GroupInfo`: `Name`, `Description`, `Commands`, `Hidden`, `LLMInstructions`

## Flag Tags

```go
type Flags struct {
    Verbose bool   `flag:"verbose" short:"v" help:"Enable verbose output"`
    Output  string `flag:"output" default:"out.txt" help:"Output path"`
    Rate    int    `flag:"rate" help:"Requests per second"`
}
```

Supported struct tags:

- `flag:"name"` - long flag name. Defaults to lowercased field name.
- `short:"x"` - single-character alias.
- `help:"text"` - help text for auto-generation.
- `default:"value"` - default if flag not provided.
- `port:"min-max"` - range validation for `Port` fields.
- `pos:"N"` - positional argument schema (see below).

## Supported Flag Types

- `string`, `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `time.Duration`
- `url.URL`, `*url.URL`
- `yargs.Port` (uint16 alias with optional range validation)
- Pointers to any of the above (for optional flags)
- `[]string` (comma-separated or repeated flags)

### Optional flags via pointers

```go
type Flags struct {
    Token *string `flag:"token" help:"Optional auth token"`
}

// Token is nil if not provided.
```

### Port validation

```go
type Flags struct {
    HTTPPort  yargs.Port  `flag:"http" port:"1-65535" help:"HTTP port"`
    AdminPort *yargs.Port `flag:"admin" port:"8000-9000" help:"Admin port"`
}
```

## Positional Argument Schemas

Define positional arguments using `pos:"N"` tags. Yargs will validate counts
and populate the struct.

```go
type Args struct {
    Service string   `pos:"0" help:"Service name"`
    Image   string   `pos:"1" help:"Image or payload"`
    Tags    []string `pos:"2*" help:"Optional tags"`
}
```

Positional tag variants:

- `pos:"0"` required argument at index 0
- `pos:"0?"` optional argument at index 0
- `pos:"0*"` variadic (0 or more) at index 0
- `pos:"0+"` variadic (1 or more) at index 0

## Aliases

Use `Aliases` on `SubCommandInfo` to register alternative command names.
`ApplyAliases` will rewrite the args to canonical names and is used by the
built-in dispatchers.

```go
config := yargs.HelpConfig{
    SubCommands: map[string]yargs.SubCommandInfo{
        "status": {Name: "status", Aliases: []string{"st", "stat"}},
    },
}

args := yargs.ApplyAliases(os.Args[1:], config)
```

Aliases also work for grouped commands (group-local aliases).

## Partial Parsing (Known Flags)

When you only want a subset of flags and want to preserve unknown args:

```go
type Flags struct {
    Host string   `flag:"host"`
    Tags []string `flag:"tags" short:"t"`
}

res, err := yargs.ParseKnownFlags[Flags](os.Args[1:], yargs.KnownFlagsOptions{
    SplitCommaSlices: true,
})
// res.Flags contains only known flags; res.RemainingArgs preserves the rest.
```

Or use the lower-level `ConsumeFlagsBySpec`:

```go
specs := map[string]yargs.ConsumeSpec{
    "host": {Kind: reflect.String},
    "tags": {Kind: reflect.Slice, SplitComma: true},
}
remaining, values := yargs.ConsumeFlagsBySpec(os.Args[1:], specs)
```

## Registry & Introspection

Use `Registry` for schema-aware command resolution and positional metadata.

```go
reg := yargs.Registry{
    Command: yargs.CommandInfo{Name: "app"},
    SubCommands: map[string]yargs.CommandSpec{
        "run": {
            Info:       yargs.SubCommandInfo{Name: "run"},
            ArgsSchema: RunArgs{},
        },
    },
}

resolved, ok, err := yargs.ResolveCommandWithRegistry(os.Args[1:], reg)
if ok {
    if spec, ok := resolved.PArg(0); ok {
        fmt.Printf("arg0 name=%s required=%v\n", spec.Name, spec.Required)
    }
}
```

## Remaining Args (`--`)

Everything after `--` is preserved in `RemainingArgs`.

```go
result, _ := yargs.ParseFlags[Flags]([]string{"-v", "arg1", "--", "--raw", "x"})
// result.Args == []string{"arg1"}
// result.RemainingArgs == []string{"--raw", "x"}
```

## Errors

Yargs exposes structured errors for control flow and diagnostics:

- `ErrHelp`, `ErrSubCommandHelp`, `ErrHelpLLM` for help requests.
- `ErrShown` from `ParseAndHandleHelp` when it already printed output.
- `InvalidFlagError` for unknown flags.
- `InvalidArgsError` for bad positional args.
- `FlagValueError` for type conversion or validation issues.

`ParseAndHandleHelp` will print a clean message for users, and if the global
flags struct has a `Verbose bool` field set to true, it will also print the full
error chain for `FlagValueError`.

## API Reference

For full, generated API docs, see:

https://pkg.go.dev/github.com/shayne/yargs

## Development

This repo uses `mise` for tool and task management.

```bash
mise install
mise run test
```

### Common tasks

- `mise run fmt`
- `mise run tidy`
- `mise run test`
- `mise run check`

## License

MIT. See `LICENSE`.
