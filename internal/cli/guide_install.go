package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sgnd/toolips/internal/out"
	"github.com/sgnd/toolips/skills"
)

// skillDirs are the folders AI agents read skills from: Claude Code reads
// .claude/skills; Codex, Cursor, Gemini CLI and others read .agents/skills.
var skillDirs = []string{".claude/skills", ".agents/skills"}

type skillFile struct {
	Path   string `json:"path"`
	Action string `json:"action"` // created | updated | unchanged
	Backup string `json:"backup,omitempty"`
}

type skillResult struct {
	Files []skillFile `json:"files"`
}

func (r skillResult) Text() string {
	var b strings.Builder
	changed := false
	for _, f := range r.Files {
		b.WriteString(f.Action + "  " + out.SafeText(f.Path) + "\n")
		if f.Backup != "" {
			b.WriteString("  old version kept as " + out.SafeText(f.Backup) + "\n")
		}
		changed = changed || f.Action != "unchanged"
	}
	if changed && !out.P.DryRun {
		b.WriteString("Restart your AI agent so it sees the skill.\n")
	}
	return b.String()
}

func guideInstallCmd() *cobra.Command {
	var project bool
	c := &cobra.Command{
		Use:   "install",
		Short: "Add the toolips skill to Claude Code, Codex, Cursor, Gemini CLI",
		Long: `Teaches Claude Code, Codex, Cursor and Gemini CLI to use toolips.
Then restart the agent.

It adds the toolips skill (toolips/SKILL.md) to the folders AI agents read
skills from: ~/.claude/skills (Claude Code) and ~/.agents/skills (Codex,
Cursor, Gemini CLI and others). The agent then uses toolips by itself when a
task needs a PDF, image, checksum, QR code, JWT, TLS, header or DNS check.

It writes only those two files and sends nothing anywhere. Safe to run
again, for example after an update: a changed file is kept as .bak.`,
		Example: `  toolips guide install             # for every project on this computer
  toolips guide install --project   # only for the folder you are in
  toolips guide install --dry-run   # show where it would go`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			base, err := os.UserHomeDir()
			if project {
				base, err = os.Getwd()
			}
			if err != nil {
				return out.Errf(out.ExitError, "NO_FOLDER", "set HOME, or run it from the project folder", "cannot find the folder to install into: %v", err)
			}
			dry := out.P.DryRun
			var paths []string
			for _, d := range skillDirs {
				paths = append(paths, filepath.Join(base, d, "toolips", "SKILL.md"))
			}
			if project {
				if err := noLinksUnder(base, paths); err != nil {
					return out.Errf(out.ExitError, "WRITE_FAILED", "remove the link, or install for your home folder without --project", "%v", err)
				}
			}
			res := skillResult{Files: []skillFile{}}
			for _, p := range paths {
				f, err := installSkill(p, dry)
				if err != nil {
					return out.Errf(out.ExitError, "WRITE_FAILED", "check that you may write to this folder", "%v", err)
				}
				res.Files = append(res.Files, f)
			}
			return done(cmd, res)
		},
	}
	c.Flags().BoolVar(&project, "project", false, "install into the current folder instead of your home folder")
	return c
}

// installSkill writes the skill to path. A different old copy is kept as a
// backup first. Every write goes through a temporary file of a random name
// in the same folder, so no name toolips writes to is one a link could hold.
func installSkill(path string, dryRun bool) (skillFile, error) {
	f := skillFile{Path: path, Action: "created"}
	old, err := os.ReadFile(path)
	switch {
	case err == nil && string(old) == skills.Toolips:
		f.Action = "unchanged"
		return f, nil
	case err == nil:
		f.Action, f.Backup = "updated", backupName(path, time.Now())
	case !errors.Is(err, fs.ErrNotExist):
		return f, err
	}
	if dryRun {
		return f, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return f, err
	}
	if f.Backup != "" {
		if err := writeAtomic(f.Backup, old, 0o600, false); err != nil {
			return f, err
		}
	}
	return f, writeAtomic(path, []byte(skills.Toolips), 0o644, true)
}

// backupName is PATH.bak, or PATH.bak.<UTC time> when that is taken. A file
// already there (the person's own .bak, or a link) is never written over.
func backupName(path string, now time.Time) string {
	name := path + ".bak"
	if _, err := os.Lstat(name); err == nil {
		name = path + ".bak." + now.UTC().Format("20060102T150405Z")
	}
	return name
}

// writeAtomic writes b next to path and renames it into place. With
// replace false it refuses a name that exists by then.
func writeAtomic(path string, b []byte, mode os.FileMode, replace bool) error {
	if !replace {
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("%s appeared while toolips wrote it; nothing was changed, run it again", path)
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// noLinksUnder refuses a link on the way from base to any of paths. A
// project folder is often a repository someone else wrote, and a link
// committed there could point the write at a file elsewhere.
func noLinksUnder(base string, paths []string) error {
	for _, p := range paths {
		rel, err := filepath.Rel(base, p)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("%s is not in %s", p, base)
		}
		cur := base
		for _, part := range strings.Split(rel, string(filepath.Separator)) {
			cur = filepath.Join(cur, part)
			fi, err := os.Lstat(cur)
			if errors.Is(err, fs.ErrNotExist) {
				break
			}
			if err != nil {
				return err
			}
			if fi.Mode()&(fs.ModeSymlink|fs.ModeIrregular) != 0 { // Irregular: a Windows junction
				return fmt.Errorf("%s is a link, and toolips does not follow links in a project folder; nothing was changed", cur)
			}
		}
	}
	return nil
}
