package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// metaCommands are listed in the docs as "about toolips" commands, not as
// tools, so the docs lists below leave them out.
var metaCommands = map[string]bool{"version": true, "list": true, "completion": true, "help": true}

// leafCommands is "noun verb" for every command the binary has.
func leafCommands() []string {
	var got []string
	for _, n := range newRoot().Commands() {
		if n.Hidden || metaCommands[n.Name()] {
			continue
		}
		for _, v := range n.Commands() {
			if !v.Hidden && v.Name() != "help" {
				got = append(got, n.Name()+" "+v.Name())
			}
		}
	}
	sort.Strings(got)
	return got
}

func repoFile(t *testing.T, parts ...string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(append([]string{"..", ".."}, parts...)...))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func sameSet(t *testing.T, where string, got, want []string) {
	t.Helper()
	sort.Strings(got)
	got = slices.Compact(got)
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Errorf("%s does not list `toolips %s`, but the binary has it", where, w)
		}
	}
	for _, g := range got {
		if !slices.Contains(want, g) {
			t.Errorf("%s lists `toolips %s`, but the binary has no such command", where, g)
		}
	}
}

// commandsSection returns the lines of the "## Commands" section of a file.
func commandsSection(t *testing.T, body string) string {
	t.Helper()
	i := strings.Index(body, "\n## Commands\n")
	if i < 0 {
		t.Fatal("no \"## Commands\" section")
	}
	rest := body[i+len("\n## Commands\n"):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

var commandLine = regexp.MustCompile("(?m)^- `toolips ([a-z-]+ [a-z-]+)`")

func listedCommands(t *testing.T, file, body string) []string {
	t.Helper()
	var got []string
	for _, m := range commandLine.FindAllStringSubmatch(commandsSection(t, body), -1) {
		got = append(got, m[1])
	}
	if len(got) == 0 {
		t.Fatalf("%s: no command lines in its Commands section", file)
	}
	return got
}

func TestCatalogVerbsMatchCommands(t *testing.T) {
	var got []string
	for _, e := range catalog {
		if metaCommands[e.Noun] {
			continue
		}
		for _, v := range e.Verbs {
			got = append(got, e.Noun+" "+v)
		}
	}
	sameSet(t, "`toolips list`", got, leafCommands())
}

func TestSkillListsEveryCommand(t *testing.T) {
	sameSet(t, "skills/toolips/SKILL.md", listedCommands(t, "SKILL.md", repoFile(t, "skills", "toolips", "SKILL.md")), leafCommands())
}

func TestAgentsListsEveryCommand(t *testing.T) {
	sameSet(t, "AGENTS.md", listedCommands(t, "AGENTS.md", repoFile(t, "AGENTS.md")), leafCommands())
}

// TestReadmeListsEveryCommand reads the command table: a first cell such as
// `pdf merge / split` stands for `pdf merge` and `pdf split`.
func TestReadmeListsEveryCommand(t *testing.T) {
	var got []string
	for _, line := range strings.Split(commandsSection(t, repoFile(t, "README.md")), "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 3 || !strings.HasPrefix(strings.TrimSpace(cells[1]), "`") {
			continue
		}
		parts := strings.Fields(strings.ReplaceAll(cells[1], "`", ""))
		if len(parts) < 2 || metaCommands[parts[0]] {
			continue
		}
		noun := parts[0]
		for _, v := range strings.Split(strings.Join(parts[1:], " "), " / ") {
			got = append(got, noun+" "+strings.TrimSpace(v))
		}
	}
	sameSet(t, "README.md", got, leafCommands())
}

// A command nobody can find from `toolips list` or the guide is a drift too.
func TestEveryCommandHasHelp(t *testing.T) {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, s := range c.Commands() {
			if s.Hidden || s.Name() == "help" || s.Name() == "completion" {
				continue
			}
			if s.Short == "" {
				t.Errorf("%s has no short help", s.CommandPath())
			}
			walk(s)
		}
	}
	walk(newRoot())
}
