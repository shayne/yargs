package yargs

import (
	"slices"
	"strings"
	"testing"
)

func TestComplete(t *testing.T) {
	config := HelpConfig{
		SubCommands: map[string]SubCommandInfo{"list": {Aliases: []string{"ls"}}, "internal": {Hidden: true}},
		Groups:      map[string]GroupInfo{"vm": {Commands: map[string]SubCommandInfo{"set": {Aliases: []string{"s"}}, "internal": {Hidden: true}}}, "secret": {Hidden: true}},
	}
	type global struct {
		Host    *string `flag:"host" short:"H"`
		Verbose bool    `short:"v"`
	}
	opts := CompletionOptions{
		Flags: func(path []string) map[string]bool {
			switch strings.Join(path, " ") {
			case "":
				return CompletionFlags(global{})
			case "vm set":
				return map[string]bool{"--size": true}
			}
			return nil
		},
		Argument: func(path []string, i int, prefix string) []string {
			if strings.Join(path, " ") == "vm set" && i == 1 {
				return []string{"my-vm", "alt", "my-vm"}
			}
			return nil
		},
	}
	for _, tc := range []struct{ words, want []string }{
		{nil, []string{"list", "ls", "vm"}},
		{[]string{"-H", "test", "v"}, []string{"vm"}},
		{[]string{"--host=lab", "-v", "vm", ""}, []string{"s", "set"}},
		{[]string{"vm", "s", "--s"}, []string{"--size"}},
		{[]string{"vm", "set", "--size", "--"}, nil},
		{[]string{"vm", "s", "--size", "2", "one", "my"}, []string{"my-vm"}},
		{[]string{"vm", "set", "--", "--"}, nil},
		{[]string{"ls", "--he"}, []string{"--help"}},
		{[]string{"wrong", ""}, nil},
		{[]string{"secret", ""}, nil},
	} {
		if got := config.Complete(tc.words, opts); !slices.Equal(got, tc.want) {
			t.Errorf("Complete(%q) = %q, want %q", tc.words, got, tc.want)
		}
	}
}
