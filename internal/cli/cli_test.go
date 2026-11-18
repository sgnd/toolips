package cli

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sgnd/toolips/internal/out"
)

func TestExecuteVersionList(t *testing.T) {
	Version = "test"
	code := runCaptured(t, []string{"version", "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	code = runCaptured(t, []string{"list", "-o", "json"})
	if code != 0 {
		t.Fatalf("list exit %d", code)
	}
}

func TestImageResizeDryRun(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.png")
	writePNG(t, src, 30, 30)
	code := runCaptured(t, []string{"image", "resize", src, "--width", "10", "--dry-run", "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
}

func TestImageBatchGlobAndDir(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "a.png"), 16, 16)
	writePNG(t, filepath.Join(dir, "b.png"), 16, 16)
	_ = os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("nope"), 0o644)

	outDir := filepath.Join(dir, "out")
	code := runCaptured(t, []string{
		"image", "grayscale", filepath.Join(dir, "*.png"),
		"--outdir", outDir, "-o", "json",
	})
	if code != 0 {
		t.Fatalf("glob exit %d", code)
	}
	ents, err := os.ReadDir(outDir)
	if err != nil || len(ents) != 2 {
		t.Fatalf("glob outdir=%v err=%v", ents, err)
	}

	outDir2 := filepath.Join(dir, "out2")
	code = runCaptured(t, []string{"image", "grayscale", dir, "--outdir", outDir2, "-o", "json"})
	if code != 0 {
		t.Fatalf("dir exit %d", code)
	}
	ents, err = os.ReadDir(outDir2)
	if err != nil || len(ents) != 2 {
		t.Fatalf("dir outdir=%v err=%v", ents, err)
	}
}

func TestImageVerbsWrite(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	writePNG(t, src, 32, 24)

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"compress", []string{"image", "compress", src, "--quality", "60", "--out", filepath.Join(dir, "c.jpg"), "-o", "json"}, filepath.Join(dir, "c.jpg")},
		{"convert", []string{"image", "convert", src, "--format", "gif", "--out", filepath.Join(dir, "c.gif"), "-o", "json"}, filepath.Join(dir, "c.gif")},
		{"crop", []string{"image", "crop", src, "--x", "2", "--y", "2", "--width", "10", "--height", "8", "--out", filepath.Join(dir, "crop.png"), "-o", "json"}, filepath.Join(dir, "crop.png")},
		{"rotate", []string{"image", "rotate", src, "--degrees", "180", "--out", filepath.Join(dir, "rot.png"), "-o", "json"}, filepath.Join(dir, "rot.png")},
		{"watermark", []string{"image", "watermark", src, "--text", "toolips", "--out", filepath.Join(dir, "wm.png"), "-o", "json"}, filepath.Join(dir, "wm.png")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code := runCaptured(t, tc.args)
			if code != 0 {
				t.Fatalf("exit %d", code)
			}
			if _, err := os.Stat(tc.want); err != nil {
				t.Fatalf("missing output: %v", err)
			}
		})
	}
}

func TestSecretScanCLI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.env")
	_ = os.WriteFile(path, []byte("KEY=AKIAIOSFODNN7EXAMPLE\n"), 0o644)
	code := runCaptured(t, []string{"secret", "scan", dir, "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
}

func TestPDFCLIPack(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "a.png")
	writePNG(t, img, 24, 24)

	pdf1 := filepath.Join(dir, "one.pdf")
	pdf2 := filepath.Join(dir, "two.pdf")
	code := runCaptured(t, []string{"pdf", "from-image", img, "--out", pdf1, "-o", "json"})
	if code != 0 {
		t.Fatalf("from-image exit %d", code)
	}
	code = runCaptured(t, []string{"pdf", "to-pdf", img, "--out", pdf2, "-o", "json"})
	if code != 0 {
		t.Fatalf("to-pdf alias exit %d", code)
	}

	merged := filepath.Join(dir, "merged.pdf")
	code = runCaptured(t, []string{"pdf", "merge", pdf1, pdf2, "--out", merged, "-o", "json"})
	if code != 0 {
		t.Fatalf("merge exit %d", code)
	}
	if st, err := os.Stat(merged); err != nil || st.Size() == 0 {
		t.Fatalf("merged missing: %v", err)
	}

	code = runCaptured(t, []string{"pdf", "merge", pdf1, pdf2, "--out", filepath.Join(dir, "dry.pdf"), "--dry-run", "-o", "json"})
	if code != 0 {
		t.Fatalf("merge dry-run exit %d", code)
	}

	splitDir := filepath.Join(dir, "pages")
	code = runCaptured(t, []string{"pdf", "split", merged, "--outdir", splitDir, "--span", "1", "-o", "json"})
	if code != 0 {
		t.Fatalf("split exit %d", code)
	}
	ents, err := os.ReadDir(splitDir)
	if err != nil || len(ents) < 2 {
		t.Fatalf("split out=%v err=%v", ents, err)
	}

	compressed := filepath.Join(dir, "c.pdf")
	code = runCaptured(t, []string{"pdf", "compress", merged, "--out", compressed, "-o", "json"})
	if code != 0 {
		t.Fatalf("compress exit %d", code)
	}

	rotated := filepath.Join(dir, "r.pdf")
	code = runCaptured(t, []string{"pdf", "rotate", pdf1, "--degrees", "90", "--out", rotated, "-o", "json"})
	if code != 0 {
		t.Fatalf("rotate exit %d", code)
	}

	wm := filepath.Join(dir, "wm.pdf")
	code = runCaptured(t, []string{"pdf", "watermark", pdf1, "--text", "DRAFT", "--out", wm, "-o", "json"})
	if code != 0 {
		t.Fatalf("watermark exit %d", code)
	}

	code = runCaptured(t, []string{"pdf", "watermark", pdf1, "-o", "json"})
	if code != out.ExitUsage {
		t.Fatalf("watermark without --text want exit %d got %d", out.ExitUsage, code)
	}

	imgDir := filepath.Join(dir, "imgs")
	code = runCaptured(t, []string{"pdf", "to-image", pdf1, "--outdir", imgDir, "-o", "json"})
	if code != 0 {
		t.Fatalf("to-image exit %d", code)
	}
}

func runCaptured(t *testing.T, args []string) int {
	t.Helper()
	var buf bytes.Buffer
	old := out.P
	out.P = &out.Printer{Format: "json", Version: Version, Start: time.Now(), W: &buf}
	defer func() { out.P = old }()
	code := executeArgs(args)
	if code == 0 && buf.Len() > 0 {
		var env out.Envelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json: %v body=%s", err, buf.String())
		}
		if !env.OK {
			t.Fatalf("envelope not ok: %s", buf.String())
		}
	}
	return code
}

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{10, 20, 30, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestMissingInputIsNotFound(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.pdf")
	cases := [][]string{
		{"pdf", "merge", missing, missing, "-o", "json"},
		{"pdf", "split", missing, "--dry-run", "-o", "json"},
		{"pdf", "compress", missing, "-o", "json"},
		{"pdf", "from-image", missing, "-o", "json"},
		{"secret", "scan", missing, "-o", "json"},
	}
	for _, args := range cases {
		if code := runCaptured(t, args); code != out.ExitNotFound {
			t.Errorf("%v: exit %d, want %d", args, code, out.ExitNotFound)
		}
	}
}

func TestPDFVerbsWrite(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "p.png")
	writePNG(t, img, 40, 40)
	a := filepath.Join(dir, "a.pdf")
	b := filepath.Join(dir, "b.pdf")
	for _, p := range []string{a, b} {
		if code := runCaptured(t, []string{"pdf", "from-image", img, img, "--out", p, "-o", "json"}); code != 0 {
			t.Fatalf("from-image exit %d", code)
		}
	}

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"merge", []string{"pdf", "merge", a, b, "--out", filepath.Join(dir, "m.pdf")}, filepath.Join(dir, "m.pdf")},
		{"split", []string{"pdf", "split", a, "--outdir", filepath.Join(dir, "split")}, filepath.Join(dir, "split")},
		{"compress", []string{"pdf", "compress", a, "--out", filepath.Join(dir, "c.pdf")}, filepath.Join(dir, "c.pdf")},
		{"rotate", []string{"pdf", "rotate", a, "--degrees", "180", "--out", filepath.Join(dir, "r.pdf")}, filepath.Join(dir, "r.pdf")},
		{"watermark", []string{"pdf", "watermark", a, "--text", "DRAFT", "--out", filepath.Join(dir, "w.pdf")}, filepath.Join(dir, "w.pdf")},
		{"to-image", []string{"pdf", "to-image", a, "--outdir", filepath.Join(dir, "imgs")}, filepath.Join(dir, "imgs")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dryArgs := append(append([]string{}, c.args...), "--dry-run", "-o", "json")
			if code := runCaptured(t, dryArgs); code != 0 {
				t.Fatalf("dry-run exit %d", code)
			}
			if _, err := os.Stat(c.want); err == nil {
				t.Fatalf("dry-run wrote %s", c.want)
			}
			if code := runCaptured(t, append(c.args, "-o", "json")); code != 0 {
				t.Fatalf("exit %d", code)
			}
			if _, err := os.Stat(c.want); err != nil {
				t.Fatalf("missing output %s: %v", c.want, err)
			}
		})
	}
}

// runFailure runs args and returns the exit code and the error envelope.
func runFailure(t *testing.T, args []string) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	old := out.P
	out.P = &out.Printer{Format: "json", Version: Version, Start: time.Now(), W: &buf}
	defer func() { out.P = old }()
	code := executeArgs(args)
	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("%v: json: %v body=%s", args, err, buf.String())
	}
	return code, env
}

func closedPort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	_ = ln.Close()
	return port
}

func TestErrorEnvelope(t *testing.T) {
	port := closedPort(t)
	cases := []struct {
		args []string
		exit int
		code string
	}{
		{[]string{"bogus"}, out.ExitUsage, "USAGE"},
		{[]string{"pdf", "merge", "only-one.pdf"}, out.ExitUsage, "USAGE"},
		{[]string{"image", "resize", "--nope"}, out.ExitUsage, "USAGE"},
		{[]string{"pdf", "nope"}, out.ExitUsage, "USAGE"},
		{[]string{"completion", "nosh"}, out.ExitUsage, "USAGE"},
		{[]string{"tls", "check", "127.0.0.1", "--port", port, "--timeout", "2s"}, out.ExitNetwork, "TLS_CHECK"},
		{[]string{"headers", "check", "http://127.0.0.1:" + port}, out.ExitNetwork, "HEADERS_CHECK"},
		{[]string{"ct", "search", "example.com", "--api", "http://127.0.0.1:" + port + "/"}, out.ExitNetwork, "CT_SEARCH"},
	}
	for _, c := range cases {
		code, env := runFailure(t, append(c.args, "-o", "json"))
		if code != c.exit {
			t.Errorf("%v: exit %d, want %d", c.args, code, c.exit)
		}
		e, _ := env["error"].(map[string]any)
		if env["ok"] != false || e == nil || e["code"] != c.code || e["hint"] == "" {
			t.Errorf("%v: envelope %v, want ok=false code=%s with hint", c.args, env, c.code)
		}
		if _, ok := env["meta"].(map[string]any); !ok {
			t.Errorf("%v: no meta", c.args)
		}
	}
}

func TestPDFErrorIsRootCause(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "bad.pdf")
	if err := os.WriteFile(bad, []byte("not a pdf"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, env := runFailure(t, []string{"pdf", "compress", bad, "-o", "json"})
	e, _ := env["error"].(map[string]any)
	msg, _ := e["message"].(string)
	if code != out.ExitError || e["code"] != "PDF_COMPRESS" || strings.Contains(msg, ": ") || msg == "" {
		t.Errorf("exit %d, error %v", code, e)
	}
}

func TestListMatchesCommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newRoot().Commands() {
		if !c.Hidden && c.Name() != "help" {
			names[c.Name()] = true
		}
	}
	for _, e := range catalog {
		if !names[e.Noun] {
			t.Errorf("list has %q but no such command", e.Noun)
		}
		delete(names, e.Noun)
	}
	for n := range names {
		t.Errorf("command %q is missing from list", n)
	}
}

// TestSkillCommandsExist keeps skills/toolips/SKILL.md honest: every
// `toolips NOUN VERB` it names must be a real command.
func TestSkillCommandsExist(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "skills", "toolips", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("toolips ([a-z-]+)(?: ([a-z-]+))?")
	root := newRoot()
	found := 0
	for _, m := range re.FindAllStringSubmatch(string(body), -1) {
		if m[1] == "is" || m[1] == "never" || m[1] == "version" && m[2] == "" {
			continue
		}
		args := []string{m[1]}
		if m[2] != "" {
			args = append(args, m[2])
		}
		cmd, rest, err := root.Find(args)
		if err != nil || cmd == root || (m[2] != "" && len(rest) > 0 && cmd.HasSubCommands()) {
			t.Errorf("SKILL.md names %q, which is not a command", m[0])
		}
		found++
	}
	if found < 10 {
		t.Fatalf("only %d commands found in SKILL.md; is the pattern still right?", found)
	}
}

func TestRefuseOverwriteUnlessForce(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "p.png")
	writePNG(t, img, 8, 8)
	pdf := filepath.Join(dir, "out.pdf")
	if code := runCaptured(t, []string{"pdf", "from-image", img, "--out", pdf}); code != 0 {
		t.Fatalf("first write exit %d", code)
	}
	for _, extra := range [][]string{nil, {"--dry-run"}} {
		args := append([]string{"pdf", "from-image", img, "--out", pdf, "-o", "json"}, extra...)
		code, env := runFailure(t, args)
		e, _ := env["error"].(map[string]any)
		if code != out.ExitError || e["code"] != "EXISTS" {
			t.Errorf("%v: exit %d, error %v", extra, code, e)
		}
	}
	if code := runCaptured(t, []string{"pdf", "from-image", img, "--out", pdf, "--force"}); code != 0 {
		t.Fatalf("--force exit %d", code)
	}

	// A batch with one existing output writes nothing at all.
	outDir := filepath.Join(dir, "gray")
	writePNG(t, filepath.Join(dir, "q.png"), 8, 8)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	taken := filepath.Join(outDir, "q.gray.png")
	writePNG(t, taken, 2, 2)
	code, _ := runFailure(t, []string{"image", "grayscale", img, filepath.Join(dir, "q.png"), "--outdir", outDir, "-o", "json"})
	if code != out.ExitError {
		t.Fatalf("batch exit %d", code)
	}
	if ents, _ := os.ReadDir(outDir); len(ents) != 1 {
		t.Errorf("batch wrote %d files before stopping, want none", len(ents)-1)
	}
}

func TestPDFInfo(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "p.png")
	writePNG(t, img, 20, 10)
	pdf := filepath.Join(dir, "a.pdf")
	if code := runCaptured(t, []string{"pdf", "from-image", img, img, "--out", pdf}); code != 0 {
		t.Fatalf("from-image exit %d", code)
	}
	var buf bytes.Buffer
	old := out.P
	out.P = &out.Printer{Format: "json", Version: Version, Start: time.Now(), W: &buf}
	code := executeArgs([]string{"pdf", "info", pdf, "-o", "json"})
	out.P = old
	if code != 0 {
		t.Fatalf("exit %d: %s", code, buf.String())
	}
	var env struct {
		Data struct {
			Files []struct {
				Pages     int    `json:"pages"`
				PageSize  string `json:"page_size"`
				Encrypted bool   `json:"encrypted"`
			} `json:"files"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil || len(env.Data.Files) != 1 {
		t.Fatalf("decode: %v %s", err, buf.String())
	}
	if f := env.Data.Files[0]; f.Pages != 2 || f.PageSize == "" || f.Encrypted {
		t.Errorf("info = %+v", f)
	}
}

func TestPDFProtectUnlock(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "p.png")
	writePNG(t, img, 8, 8)
	pdf := filepath.Join(dir, "a.pdf")
	if code := runCaptured(t, []string{"pdf", "from-image", img, "--out", pdf}); code != 0 {
		t.Fatalf("from-image exit %d", code)
	}
	pw := filepath.Join(dir, "pw.txt")
	bad := filepath.Join(dir, "bad.txt")
	_ = os.WriteFile(pw, []byte("open sesame\n"), 0o600)
	_ = os.WriteFile(bad, []byte("wrong\n"), 0o600)
	locked := filepath.Join(dir, "locked.pdf")

	if code, _ := runFailure(t, []string{"pdf", "protect", pdf, "--out", locked, "-o", "json"}); code != out.ExitUsage {
		t.Errorf("no --password-file: exit %d", code)
	}
	if code := runCaptured(t, []string{"pdf", "protect", pdf, "--password-file", pw, "--out", locked, "-o", "json"}); code != 0 {
		t.Fatalf("protect exit %d", code)
	}
	if code, _ := runFailure(t, []string{"pdf", "unlock", locked, "--password-file", bad, "--out", filepath.Join(dir, "x.pdf"), "-o", "json"}); code != out.ExitError {
		t.Errorf("wrong password: exit %d", code)
	}
	if code := runCaptured(t, []string{"pdf", "unlock", locked, "--password-file", pw, "--out", filepath.Join(dir, "open.pdf"), "-o", "json"}); code != 0 {
		t.Errorf("unlock exit %d", code)
	}
}
