// Copyright (c) 2026 AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yargs

import (
	"slices"
	"strings"
)

// CompletionOptions supplies flags and optional positional values for each command.
// Flags maps each spelling (e.g. "--host" or "-H") to whether it takes a value.
// The empty path denotes global flags, which are available throughout the command.
type CompletionOptions struct {
	Flags    func(path []string) map[string]bool
	Argument func(path []string, position int, prefix string) []string
}

// CompletionFlags returns the long and short flag spellings in a tagged struct.
func CompletionFlags(schema any) map[string]bool {
	flags := make(map[string]bool)
	for _, f := range agentFlagInfo(schema) {
		takesValue := strings.TrimLeft(f.Type, "*") != "bool"
		flags["--"+f.Name] = takesValue
		if f.ShortName != "" {
			flags["-"+f.ShortName] = takesValue
		}
	}
	return flags
}

// Complete returns sorted candidates for the word at the cursor. Pass the words
// after the executable through the cursor, including an empty last word after a space.
// Completing a command never invokes its handler.
func (config HelpConfig) Complete(words []string, opts CompletionOptions) []string {
	if len(words) == 0 {
		words = []string{""}
	}
	var path []string
	flags := map[string]bool{"--help": false, "-h": false}
	addFlags := func() {
		if opts.Flags != nil {
			for name, takesValue := range opts.Flags(path) {
				flags[name] = takesValue
			}
		}
	}
	addFlags()
	commands := config.SubCommands
	waitingValue, selected := false, false
	position := 0
	for _, word := range words[:len(words)-1] {
		if waitingValue {
			waitingValue = false
			continue
		}
		if word == "--" {
			return nil
		}
		if strings.HasPrefix(word, "-") {
			name, _, attached := strings.Cut(word, "=")
			waitingValue = flags[name] && !attached
			continue
		}
		if selected {
			position++
			continue
		}
		if group, ok := config.Groups[word]; len(path) == 0 && ok && !group.Hidden {
			path = []string{word}
			commands = group.Commands
			addFlags()
			continue
		}
		found := false
		for name, command := range commands {
			if !command.Hidden && (name == word || slices.Contains(command.Aliases, word)) {
				path = append(path, name)
				selected, found = true, true
				addFlags()
				break
			}
		}
		if !found {
			return nil
		}
	}

	prefix := words[len(words)-1]
	var candidates []string
	switch {
	case waitingValue:
		return nil
	case strings.HasPrefix(prefix, "-"):
		for name := range flags {
			candidates = append(candidates, name)
		}
	case selected:
		if opts.Argument != nil {
			candidates = opts.Argument(path, position, prefix)
		}
	default:
		if len(path) == 0 {
			for name, group := range config.Groups {
				if !group.Hidden {
					candidates = append(candidates, name)
				}
			}
		}
		for name, command := range commands {
			if !command.Hidden {
				candidates = append(candidates, name)
				candidates = append(candidates, command.Aliases...)
			}
		}
	}
	var matches []string
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, prefix) {
			matches = append(matches, candidate)
		}
	}
	slices.Sort(matches)
	return slices.Compact(matches)
}
