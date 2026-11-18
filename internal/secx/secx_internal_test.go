package secx

import (
	"crypto/tls"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGradeAndVersionNames(t *testing.T) {
	for score, want := range map[int]string{100: "A", 80: "B", 60: "C", 30: "D", 0: "F"} {
		if got := gradeFor(score); got != want {
			t.Errorf("gradeFor(%d) = %s, want %s", score, got, want)
		}
	}
	for v, want := range map[uint16]string{tls.VersionTLS13: "TLS1.3", tls.VersionTLS12: "TLS1.2", tls.VersionTLS11: "TLS1.1", tls.VersionTLS10: "TLS1.0", 0x1234: "0x1234"} {
		if got := tlsVersionName(v); got != want {
			t.Errorf("tlsVersionName(%x) = %s, want %s", v, got, want)
		}
	}
	if truncate("abc", 5) != "abc" || truncate("abcdef", 3) != "abc…" {
		t.Error("truncate")
	}
}

func TestTextRenderers(t *testing.T) {
	texts := []string{
		TLSResult{Host: "h", Port: "443", Error: "refused"}.Text(),
		TLSResult{OK: true, Host: "h", Port: "443", Protocol: "TLS1.3"}.Text(),
		HeaderCheck{}.Text(),
		ScanResult{}.Text(),
		CTResult{}.Text(),
	}
	for i, s := range texts {
		if !strings.HasSuffix(s, "\n") && s != "" {
			t.Errorf("text %d has no trailing newline: %q", i, s)
		}
	}
	if !strings.Contains(texts[0], "FAIL") {
		t.Errorf("failed TLS text: %q", texts[0])
	}
}

func TestCheckTLSClosedPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	_ = ln.Close()
	if res := CheckTLS("127.0.0.1", port, time.Second); res.OK || res.Error == "" {
		t.Errorf("closed port: %+v", res)
	}
}

func TestScanReportsWhatItSkips(t *testing.T) {
	dir := t.TempDir()
	key := "AKIA" + strings.Repeat("A", 16)
	long := key + "\n" + strings.Repeat("x", 2<<20-100) + "\n" + key + "\n"
	big := strings.Repeat("y", 3<<20)
	files := map[string]string{"long.js": long, "big.txt": big}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	res, err := ScanPaths([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Count != 1 {
		t.Errorf("want the finding before the long line kept, got %d findings", res.Count)
	}
	reasons := map[string]string{}
	for _, s := range res.Skipped {
		reasons[filepath.Base(s.File)] = s.Reason
	}
	if !strings.Contains(reasons["long.js"], "1 MiB") || !strings.Contains(reasons["big.txt"], "2 MiB") {
		t.Errorf("skipped = %+v", res.Skipped)
	}
	if !strings.Contains(res.Text(), "not fully scanned") {
		t.Errorf("text hides skipped files: %q", res.Text())
	}
}

func TestScanTextEscapesFileNames(t *testing.T) {
	res := ScanResult{Count: 1, Findings: []Finding{{File: "a\x1b[2Jb\rc", Line: 1, Rule: "r", Snippet: "s"}}}
	if txt := res.Text(); strings.ContainsAny(txt, "\x1b\r") {
		t.Errorf("control characters reached the terminal: %q", txt)
	}
}
