package secx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CTEntry is one certificate transparency hit.
type CTEntry struct {
	IssuerName   string   `json:"issuer_name"`
	CommonName   string   `json:"common_name"`
	NameValue    string   `json:"name_value"`
	NotBefore    string   `json:"not_before"`
	NotAfter     string   `json:"not_after"`
	SerialNumber string   `json:"serial_number"`
	DNSNames     []string `json:"dns_names,omitempty"`
}

// CTResult is a CT search response.
type CTResult struct {
	Domain  string    `json:"domain"`
	Count   int       `json:"count"`
	Entries []CTEntry `json:"entries"`
	Source  string    `json:"source"`
	Error   string    `json:"error,omitempty"`
}

func (c CTResult) Text() string {
	if c.Error != "" {
		return fmt.Sprintf("%s  FAIL  %s\n", c.Domain, c.Error)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %d certificates (%s)\n", c.Domain, c.Count, c.Source)
	limit := len(c.Entries)
	if limit > 20 {
		limit = 20
	}
	for i := 0; i < limit; i++ {
		e := c.Entries[i]
		fmt.Fprintf(&b, "  %s  %s → %s  %s\n", e.CommonName, e.NotBefore, e.NotAfter, e.IssuerName)
	}
	if len(c.Entries) > limit {
		fmt.Fprintf(&b, "  … and %d more\n", len(c.Entries)-limit)
	}
	return b.String()
}

// crt.sh response shape (subset).
type crtShRow struct {
	IssuerName   string `json:"issuer_name"`
	CommonName   string `json:"common_name"`
	NameValue    string `json:"name_value"`
	NotBefore    string `json:"not_before"`
	NotAfter     string `json:"not_after"`
	SerialNumber string `json:"serial_number"`
}

const defaultCTURL = "https://crt.sh/"

// SearchCT queries a public CT API (crt.sh by default) for domain.
func SearchCT(domain string, client *http.Client, baseURL string) CTResult {
	res := CTResult{Domain: domain, Source: "crt.sh"}
	if domain == "" {
		res.Error = "domain is required"
		return res
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	if baseURL == "" {
		baseURL = defaultCTURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	q := u.Query()
	q.Set("q", domain)
	q.Set("output", "json")
	u.RawQuery = q.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		res.Error = err.Error()
		return res
	}
	if resp.StatusCode >= 400 {
		res.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return res
	}
	var rows []crtShRow
	if err := json.Unmarshal(body, &rows); err != nil {
		res.Error = "decode: " + err.Error()
		return res
	}
	seen := map[string]bool{}
	for _, row := range rows {
		key := row.SerialNumber + "|" + row.CommonName
		if seen[key] {
			continue
		}
		seen[key] = true
		names := splitNames(row.NameValue)
		res.Entries = append(res.Entries, CTEntry{
			IssuerName:   row.IssuerName,
			CommonName:   row.CommonName,
			NameValue:    row.NameValue,
			NotBefore:    row.NotBefore,
			NotAfter:     row.NotAfter,
			SerialNumber: row.SerialNumber,
			DNSNames:     names,
		})
	}
	res.Count = len(res.Entries)
	return res
}

func splitNames(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '\n' || r == ','
	})
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
