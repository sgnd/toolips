package secx

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
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

type fakeResolver struct {
	ips  []net.IPAddr
	mx   []*net.MX
	txt  map[string][]string
	fail error
}

func (f fakeResolver) LookupIPAddr(_ context.Context, _ string) ([]net.IPAddr, error) {
	if f.fail != nil {
		return nil, f.fail
	}
	if len(f.ips) == 0 {
		return nil, &net.DNSError{Err: "no such host", IsNotFound: true}
	}
	return f.ips, nil
}

func (f fakeResolver) LookupMX(_ context.Context, _ string) ([]*net.MX, error) {
	if f.fail != nil {
		return nil, f.fail
	}
	if len(f.mx) == 0 {
		return nil, &net.DNSError{Err: "no such host", IsNotFound: true}
	}
	return f.mx, nil
}

func (f fakeResolver) LookupTXT(_ context.Context, name string) ([]string, error) {
	if f.fail != nil {
		return nil, f.fail
	}
	if t, ok := f.txt[name]; ok {
		return t, nil
	}
	return nil, &net.DNSError{Err: "no such host", IsNotFound: true}
}

func TestCheckDNS(t *testing.T) {
	good := fakeResolver{
		ips: []net.IPAddr{{IP: net.ParseIP("192.0.2.1")}, {IP: net.ParseIP("2001:db8::1")}},
		mx:  []*net.MX{{Host: "mx2.example.com.", Pref: 20}, {Host: "mx1.example.com.", Pref: 10}},
		txt: map[string][]string{
			"example.com":        {"v=spf1 mx -all", "other"},
			"_dmarc.example.com": {"v=DMARC1; p=reject"},
		},
	}
	res := CheckDNS("Example.com.", good, time.Second)
	if res.Error != "" || len(res.A) != 1 || len(res.AAAA) != 1 || res.MX[0].Host != "mx1.example.com" || len(res.SPF) != 1 {
		t.Fatalf("good = %+v", res)
	}
	if len(res.Problems) != 0 {
		t.Errorf("good domain has problems: %v", res.Problems)
	}

	weak := fakeResolver{
		ips: []net.IPAddr{{IP: net.ParseIP("192.0.2.1")}},
		txt: map[string][]string{
			"example.com":        {"v=spf1 a -all", "v=spf1 mx -all"},
			"_dmarc.example.com": {"v=DMARC1; p = none"},
		},
	}
	res = CheckDNS("example.com", weak, time.Second)
	got := strings.Join(res.Problems, "|")
	for _, want := range []string{"no MX", "2 SPF records", "p=none"} {
		if !strings.Contains(got, want) {
			t.Errorf("problems %q lack %q", got, want)
		}
	}

	res = CheckDNS("nope.example", fakeResolver{}, time.Second)
	if !res.NotFound || res.Error == "" {
		t.Errorf("missing domain = %+v", res)
	}
	res = CheckDNS("example.com", fakeResolver{fail: errors.New("connection refused")}, time.Second)
	if res.NotFound || res.Error == "" {
		t.Errorf("network failure = %+v", res)
	}
	if txt := (DNSCheck{Domain: "d", Problems: []string{"x\x1b[2J"}}).Text(); strings.Contains(txt, "\x1b") {
		t.Errorf("text not escaped: %q", txt)
	}
}

func TestCheckDNSNullMX(t *testing.T) {
	r := fakeResolver{
		ips: []net.IPAddr{{IP: net.ParseIP("192.0.2.1")}},
		mx:  []*net.MX{{Host: ".", Pref: 0}},
		txt: map[string][]string{"example.com": {"v=spf1 -all"}, "_dmarc.example.com": {"v=DMARC1; p=reject"}},
	}
	res := CheckDNS("example.com", r, time.Second)
	if len(res.Problems) != 0 || !strings.Contains(res.Text(), "null MX") {
		t.Errorf("null MX: problems %v, text %q", res.Problems, res.Text())
	}
}

func jwtPart(v string) string { return base64.RawURLEncoding.EncodeToString([]byte(v)) }

func TestDecodeJWT(t *testing.T) {
	tok := jwtPart(`{"alg":"HS256","typ":"JWT"}`) + "." + jwtPart(`{"sub":"42","exp":1700000000,"iat":1690000000}`) + ".c2ln"
	now := time.Unix(1800000000, 0)
	j, err := DecodeJWT("Bearer "+tok+"\n", now)
	if err != nil {
		t.Fatal(err)
	}
	if j.Header["alg"] != "HS256" || j.Claims["sub"] != "42" || !j.Expired || !j.Signed || j.Verified {
		t.Errorf("decoded = %+v", j)
	}
	if j.Times["exp"] != "2023-11-14T22:13:20Z" {
		t.Errorf("exp = %s", j.Times["exp"])
	}
	if !strings.Contains(j.Text(), "signature not checked") {
		t.Error("text does not say the signature is not checked")
	}
	for _, bad := range []string{"", "a.b", "!!.e30.x", jwtPart("[1]") + ".e30.x"} {
		if _, err := DecodeJWT(bad, now); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
}
