package secx_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sgnd/toolips/internal/secx"
)

func TestSecretScan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leak.txt")
	content := "export AWS_KEY=AKIAIOSFODNN7EXAMPLE\ntoken=ghp_abcdefghijklmnopqrstuvwxyz0123456789\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := secx.ScanPaths([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Count < 2 {
		t.Fatalf("expected findings, got %+v", res)
	}
	for _, f := range res.Findings {
		if strings.Contains(f.Snippet, "AKIAIOSFODNN7EXAMPLE") || strings.Contains(f.Snippet, "ghp_abcdefghijklmnopqrstuvwxyz0123456789") {
			t.Fatalf("snippet leaked secret: %q", f.Snippet)
		}
		if !strings.Contains(f.Snippet, "[redacted]") {
			t.Fatalf("expected [redacted] in snippet: %q", f.Snippet)
		}
	}
}

func TestCTSearchEscapesControls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"issuer_name":"CA\u001b[31m","common_name":"evil.com","name_value":"evil.com","not_before":"2024-01-01","not_after":"2025-01-01","serial_number":"1"}]`))
	}))
	defer srv.Close()
	res := secx.SearchCT("evil.com", srv.Client(), srv.URL)
	if res.Error != "" || res.Count != 1 {
		t.Fatalf("%+v", res)
	}
	if strings.Contains(res.Entries[0].IssuerName, "\x1b") {
		t.Fatalf("raw ESC in issuer: %q", res.Entries[0].IssuerName)
	}
	if !strings.Contains(res.Entries[0].IssuerName, `\x1b`) {
		t.Fatalf("expected escaped ESC in %q", res.Entries[0].IssuerName)
	}
}

func TestHeadersCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	res := secx.CheckHeaders(srv.URL, srv.Client())
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if res.Score < 40 {
		t.Fatalf("score=%d grade=%s missing=%v", res.Score, res.Grade, res.Missing)
	}
	if len(res.Present) < 3 {
		t.Fatalf("present=%v", res.Present)
	}
}

func TestCTSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("output") != "json" {
			t.Errorf("missing output=json")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
		  {"issuer_name":"Test CA","common_name":"example.com","name_value":"example.com\nwww.example.com","not_before":"2024-01-01","not_after":"2025-01-01","serial_number":"1"},
		  {"issuer_name":"Test CA","common_name":"example.com","name_value":"example.com","not_before":"2024-01-01","not_after":"2025-01-01","serial_number":"1"}
		]`))
	}))
	defer srv.Close()

	res := secx.SearchCT("example.com", srv.Client(), srv.URL)
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if res.Count != 1 {
		t.Fatalf("dedupe failed: %+v", res)
	}
	if !strings.Contains(res.Text(), "example.com") {
		t.Fatalf("text=%q", res.Text())
	}
}

func TestCheckTLSSkipWithoutNetwork(t *testing.T) {
	if os.Getenv("TOOLIPS_NET_TEST") == "" {
		t.Skip("set TOOLIPS_NET_TEST=1 to run live TLS check")
	}
	res := secx.CheckTLS("example.com", "443", 10*time.Second)
	if !res.OK {
		t.Fatalf("%+v", res)
	}
	if res.DaysLeft < 0 {
		t.Fatalf("expired: %+v", res)
	}
}

func TestCheckTLSLocal(t *testing.T) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "toolips-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	leafCert, err := x509.ParseCertificate(leafDER)
	if err != nil {
		t.Fatal(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDER})
	keyDER, err := x509.MarshalECPrivateKey(leafKey)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{tlsCert}})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c.(*tls.Conn).Handshake()
			_ = c.Close()
		}
	}()

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	res := secx.CheckTLSWithRoots("127.0.0.1", port, 5*time.Second, roots)
	if !res.OK {
		t.Fatalf("%+v", res)
	}
	if res.Protocol == "" || res.DaysLeft < 0 {
		t.Fatalf("%+v", res)
	}
	if secx.CertSummary(leafCert) != "localhost" {
		t.Fatalf("dns=%s", secx.CertSummary(leafCert))
	}
}
