package cli

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rsc.io/qr"
)

func TestQRMakeWritesPNGOfTheRightSize(t *testing.T) {
	out := filepath.Join(t.TempDir(), "q.png")
	code, env := runFailure(t, []string{"qr", "make", "https://toolips.efolusi.com", "--out", out, "--scale", "5", "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d: %v", code, env)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	c, _ := qr.Encode("https://toolips.efolusi.com", qr.M)
	want := (c.Size + 2*4) * 5 // 4 modules of quiet zone on each side
	if b := img.Bounds(); b.Dx() != want || b.Dy() != want {
		t.Errorf("image is %dx%d, want %dx%d", b.Dx(), b.Dy(), want, want)
	}
	// the top-left finder pattern starts at module (0,0): black, inside the quiet zone offset
	if r, _, _, _ := img.At(4*5+1, 4*5+1).RGBA(); r != 0 {
		t.Error("first finder module is not black")
	}
	if r, _, _, _ := img.At(1, 1).RGBA(); r == 0 {
		t.Error("quiet zone is not white")
	}
}

func TestQRMakeLevelChangesTheCode(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.png"), filepath.Join(dir, "b.png")
	runFailure(t, []string{"qr", "make", "hello world", "--out", a, "--level", "L", "-o", "json"})
	runFailure(t, []string{"qr", "make", "hello world", "--out", b, "--level", "H", "-o", "json"})
	ba, _ := os.ReadFile(a)
	bb, _ := os.ReadFile(b)
	if len(ba) == 0 || string(ba) == string(bb) {
		t.Error("level L and H gave the same image")
	}
}

func TestQRMakeReadsStdin(t *testing.T) {
	out := filepath.Join(t.TempDir(), "s.png")
	in := filepath.Join(t.TempDir(), "in.txt")
	if err := os.WriteFile(in, []byte("from stdin\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(in)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	old := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = old }()
	if code, env := runFailure(t, []string{"qr", "make", "-", "--out", out, "-o", "json"}); code != 0 {
		t.Fatalf("exit %d: %v", code, env)
	}
	if fi, err := os.Stat(out); err != nil || fi.Size() == 0 {
		t.Errorf("no image from stdin: %v", err)
	}
}

func TestQRMakeErrors(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "e.png")
	cases := []struct {
		name string
		args []string
		exit int
		code string
	}{
		{"bad level", []string{"qr", "make", "x", "--out", out, "--level", "Z"}, 2, "USAGE"},
		{"bad scale", []string{"qr", "make", "x", "--out", out, "--scale", "0"}, 2, "USAGE"},
		{"too long", []string{"qr", "make", strings.Repeat("x", 4000), "--out", out}, 1, "QR_TOO_LONG"},
		{"empty", []string{"qr", "make", "", "--out", out}, 2, "USAGE"},
	}
	for _, c := range cases {
		code, env := runFailure(t, append(c.args, "-o", "json"))
		e, _ := env["error"].(map[string]any)
		if code != c.exit || e["code"] != c.code {
			t.Errorf("%s: exit %d code %v", c.name, code, e["code"])
		}
	}
	if _, err := os.Stat(out); err == nil {
		t.Error("a failed run left a file")
	}
}

func TestQRMakeDryRunAndOverwrite(t *testing.T) {
	out := filepath.Join(t.TempDir(), "d.png")
	if code, _ := runFailure(t, []string{"qr", "make", "x", "--out", out, "--dry-run", "-o", "json"}); code != 0 {
		t.Fatalf("dry run exit %d", code)
	}
	if _, err := os.Stat(out); err == nil {
		t.Fatal("dry run wrote the file")
	}
	runFailure(t, []string{"qr", "make", "x", "--out", out, "-o", "json"})
	code, env := runFailure(t, []string{"qr", "make", "y", "--out", out, "-o", "json"})
	if e, _ := env["error"].(map[string]any); code != 1 || e["code"] != "EXISTS" {
		t.Errorf("overwrite: exit %d %v", code, e)
	}
	if code, _ := runFailure(t, []string{"qr", "make", "y", "--out", out, "--force", "-o", "json"}); code != 0 {
		t.Errorf("--force exit %d", code)
	}
}
