package secx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/sgnd/toolips/internal/out"
	"net"
	"strings"
	"time"
)

// TLSResult is the outcome of a TLS check.
type TLSResult struct {
	Host           string   `json:"host"`
	Port           string   `json:"port"`
	OK             bool     `json:"ok"`
	ServerName     string   `json:"server_name,omitempty"`
	Protocol       string   `json:"protocol,omitempty"`
	Cipher         string   `json:"cipher,omitempty"`
	NotBefore      string   `json:"not_before,omitempty"`
	NotAfter       string   `json:"not_after,omitempty"`
	DaysLeft       int      `json:"days_left,omitempty"`
	Issuer         string   `json:"issuer,omitempty"`
	Subject        string   `json:"subject,omitempty"`
	DNSNames       []string `json:"dns_names,omitempty"`
	ChainLen       int      `json:"chain_len,omitempty"`
	VerifiedChains int      `json:"verified_chains,omitempty"`
	Error          string   `json:"error,omitempty"`
}

func (r TLSResult) Text() string {
	if !r.OK {
		return fmt.Sprintf("%s:%s  FAIL  %s\n", r.Host, r.Port, out.SafeText(r.Error))
	}
	return fmt.Sprintf("%s:%s  OK  %s  expires %s (%d days)  %s\n",
		r.Host, r.Port, r.Protocol, r.NotAfter, r.DaysLeft, out.SafeText(r.Subject))
}

// CheckTLS dials host:port and inspects the certificate.
func CheckTLS(host, port string, timeout time.Duration) TLSResult {
	return CheckTLSWithRoots(host, port, timeout, nil)
}

// CheckTLSWithRoots is like CheckTLS but trusts an optional root pool
// (used by tests with a local CA).
func CheckTLSWithRoots(host, port string, timeout time.Duration, roots *x509.CertPool) TLSResult {
	if port == "" {
		port = "443"
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	res := TLSResult{Host: host, Port: port}
	addr := net.JoinHostPort(host, port)
	dialer := &net.Dialer{Timeout: timeout}
	cfg := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
		RootCAs:    roots,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, cfg)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer conn.Close()
	state := conn.ConnectionState()
	res.OK = true
	res.ServerName = state.ServerName
	res.Protocol = tlsVersionName(state.Version)
	res.Cipher = tls.CipherSuiteName(state.CipherSuite)
	res.ChainLen = len(state.PeerCertificates)
	res.VerifiedChains = len(state.VerifiedChains)
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		res.NotBefore = cert.NotBefore.UTC().Format(time.RFC3339)
		res.NotAfter = cert.NotAfter.UTC().Format(time.RFC3339)
		res.DaysLeft = int(time.Until(cert.NotAfter).Hours() / 24)
		res.Issuer = cert.Issuer.String()
		res.Subject = cert.Subject.String()
		res.DNSNames = append([]string{}, cert.DNSNames...)
	}
	return res
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS1.3"
	case tls.VersionTLS12:
		return "TLS1.2"
	case tls.VersionTLS11:
		return "TLS1.1"
	case tls.VersionTLS10:
		return "TLS1.0"
	default:
		return fmt.Sprintf("0x%04x", v)
	}
}

// CertSummary is a short view of a leaf certificate (for tests).
func CertSummary(cert *x509.Certificate) string {
	return strings.Join(cert.DNSNames, ",")
}
