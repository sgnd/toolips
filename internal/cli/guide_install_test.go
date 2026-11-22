package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sgnd/toolips/skills"
)

func skillPaths(base string) []string {
	return []string{
		filepath.Join(base, ".claude", "skills", "toolips", "SKILL.md"),
		filepath.Join(base, ".agents", "skills", "toolips", "SKILL.md"),
	}
}

func installActions(t *testing.T, env map[string]any) []string {
	t.Helper()
	data, _ := env["data"].(map[string]any)
	files, _ := data["files"].([]any)
	var got []string
	for _, f := range files {
		m := f.(map[string]any)
		got = append(got, m["action"].(string))
	}
	return got
}

func TestGuideInstallHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	code, env := runFailure(t, []string{"guide", "install", "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d: %v", code, env)
	}
	if got := installActions(t, env); strings.Join(got, ",") != "created,created" {
		t.Fatalf("actions = %v", got)
	}
	for _, p := range skillPaths(home) {
		b, err := os.ReadFile(p)
		if err != nil || string(b) != skills.Toolips {
			t.Errorf("%s: err %v, content differs from the embedded skill", p, err)
		}
	}

	_, env = runFailure(t, []string{"guide", "install", "-o", "json"})
	if got := installActions(t, env); strings.Join(got, ",") != "unchanged,unchanged" {
		t.Errorf("second run actions = %v", got)
	}
}

func TestGuideInstallKeepsOldCopy(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	p := skillPaths(home)[0]
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("my own notes"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, env := runFailure(t, []string{"guide", "install", "-o", "json"})
	if got := installActions(t, env); strings.Join(got, ",") != "updated,created" {
		t.Fatalf("actions = %v", got)
	}
	if b, err := os.ReadFile(p + ".bak"); err != nil || string(b) != "my own notes" {
		t.Errorf("backup: %v %q", err, b)
	}
	if b, _ := os.ReadFile(p); string(b) != skills.Toolips {
		t.Error("skill not written over the old copy")
	}
}

func TestGuideInstallDryRunWritesNothing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	code, env := runFailure(t, []string{"guide", "install", "--dry-run", "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if got := installActions(t, env); strings.Join(got, ",") != "created,created" {
		t.Errorf("plan = %v", got)
	}
	if entries, _ := os.ReadDir(home); len(entries) != 0 {
		t.Errorf("dry run wrote %d entries", len(entries))
	}
}

func TestGuideInstallProject(t *testing.T) {
	home, proj := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Chdir(proj)
	if code, env := runFailure(t, []string{"guide", "install", "--project", "-o", "json"}); code != 0 {
		t.Fatalf("exit %d: %v", code, env)
	}
	for _, p := range skillPaths(proj) {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("project install missing %s", p)
		}
	}
	if entries, _ := os.ReadDir(home); len(entries) != 0 {
		t.Error("--project wrote into the home folder")
	}
}

func TestGuideInstallProjectRefusesLink(t *testing.T) {
	proj, elsewhere := t.TempDir(), t.TempDir()
	if err := os.Symlink(elsewhere, filepath.Join(proj, ".claude")); err != nil {
		t.Skip("no symlinks here")
	}
	t.Chdir(proj)
	code, env := runFailure(t, []string{"guide", "install", "--project", "-o", "json"})
	e, _ := env["error"].(map[string]any)
	if code == 0 || e["code"] != "WRITE_FAILED" {
		t.Fatalf("exit %d, error %v", code, e)
	}
	if entries, _ := os.ReadDir(elsewhere); len(entries) != 0 {
		t.Error("wrote through the link")
	}
}

func TestGuideStillPrintsPlaybook(t *testing.T) {
	code, env := runFailure(t, []string{"guide", "-o", "json"})
	data, _ := env["data"].(map[string]any)
	if code != 0 || !strings.Contains(data["guide"].(string), "toolips guide install") {
		t.Errorf("guide must point to guide install; exit %d", code)
	}
}
