package secx

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HeaderCheck grades common security headers on a URL.
type HeaderCheck struct {
	URL     string         `json:"url"`
	Grade   string         `json:"grade"`
	Score   int            `json:"score"`
	Present []HeaderStatus `json:"present"`
	Missing []string       `json:"missing"`
	Error   string         `json:"error,omitempty"`
}

// HeaderStatus is one observed header.
type HeaderStatus struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (h HeaderCheck) Text() string {
	if h.Error != "" {
		return fmt.Sprintf("%s  FAIL  %s\n", h.URL, h.Error)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s  grade %s (%d/100)\n", h.URL, h.Grade, h.Score)
	for _, p := range h.Present {
		fmt.Fprintf(&b, "  + %s: %s\n", p.Name, truncate(safeText(p.Value), 80))
	}
	for _, m := range h.Missing {
		fmt.Fprintf(&b, "  - missing %s\n", m)
	}
	return b.String()
}

var securityHeaders = []string{
	"Strict-Transport-Security",
	"Content-Security-Policy",
	"X-Content-Type-Options",
	"X-Frame-Options",
	"Referrer-Policy",
	"Permissions-Policy",
}

// CheckHeaders fetches url and grades security headers.
func CheckHeaders(url string, client *http.Client) HeaderCheck {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		}}
	}
	res := HeaderCheck{URL: url}
	if !strings.Contains(url, "://") {
		url = "https://" + url
		res.URL = url
	}
	resp, err := client.Get(url)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))

	points := 0
	per := 100 / len(securityHeaders)
	for _, name := range securityHeaders {
		v := resp.Header.Get(name)
		if v == "" {
			res.Missing = append(res.Missing, name)
			continue
		}
		res.Present = append(res.Present, HeaderStatus{Name: name, Value: safeText(v)})
		points += per
	}
	res.Score = points
	res.Grade = gradeFor(points)
	return res
}

func gradeFor(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score >= 50:
		return "C"
	case score >= 25:
		return "D"
	default:
		return "F"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
