package secx

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sgnd/toolips/internal/out"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Finding is one secret-like match.
type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Snippet string `json:"snippet"`
}

// ScanResult lists findings under paths.
type ScanResult struct {
	Findings []Finding `json:"findings"`
	Scanned  int       `json:"scanned"`
	Count    int       `json:"count"`
	// Skipped lists files not scanned in full, so "no findings" is never
	// read as "no secrets" for them.
	Skipped []Skipped `json:"skipped,omitempty"`
}

// Skipped is a file the scan did not read in full, and why.
type Skipped struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

func (s ScanResult) Text() string {
	var b strings.Builder
	if s.Count == 0 {
		fmt.Fprintf(&b, "scanned %d files, no secrets found\n", s.Scanned)
	} else {
		fmt.Fprintf(&b, "scanned %d files, %d findings\n", s.Scanned, s.Count)
	}
	for _, f := range s.Skipped {
		fmt.Fprintf(&b, "%s  not fully scanned: %s\n", out.SafeText(f.File), f.Reason)
	}
	for _, f := range s.Findings {
		fmt.Fprintf(&b, "%s:%d  %s  %s\n", out.SafeText(f.File), f.Line, f.Rule, f.Snippet)
	}
	return b.String()
}

type rule struct {
	name string
	re   *regexp.Regexp
}

var secretRules = []rule{
	{"aws_access_key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"github_token", regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{20,}`)},
	{"slack_token", regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{10,}`)},
	{"private_key", regexp.MustCompile(`-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`)},
	{"generic_api_key", regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|access[_-]?token)\s*[:=]\s*['"]?[A-Za-z0-9_\-]{16,}`)},
}

var skipExt = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".bmp": true, ".tif": true, ".tiff": true,
	".pdf": true, ".zip": true, ".gz": true, ".tar": true, ".woff": true,
	".woff2": true, ".ico": true, ".mp4": true, ".mp3": true,
}

// ScanPaths walks files and dirs for secret-like patterns.
func ScanPaths(paths []string) (ScanResult, error) {
	var res ScanResult
	for _, p := range paths {
		err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				base := info.Name()
				if base == ".git" || base == "node_modules" || base == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if info.Size() > 2<<20 {
				res.Skipped = append(res.Skipped, Skipped{File: path, Reason: "larger than 2 MiB"})
				return nil
			}
			if skipExt[strings.ToLower(filepath.Ext(path))] {
				return nil
			}
			res.Scanned++
			findings, err := scanFile(path)
			res.Findings = append(res.Findings, findings...)
			switch {
			case errors.Is(err, bufio.ErrTooLong):
				res.Skipped = append(res.Skipped, Skipped{File: path, Reason: "a line is longer than 1 MiB; read up to it"})
			case err != nil:
				res.Skipped = append(res.Skipped, Skipped{File: path, Reason: "unreadable"})
			}
			return nil
		})
		if err != nil {
			return res, err
		}
	}
	res.Count = len(res.Findings)
	return res, nil
}

func scanFile(path string) ([]Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var found []Finding
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		text := sc.Text()
		for _, r := range secretRules {
			if loc := r.re.FindStringIndex(text); loc != nil {
				found = append(found, Finding{
					File:    path,
					Line:    line,
					Rule:    r.name,
					Snippet: redactSnippet(text, loc[0], loc[1]),
				})
			}
		}
	}
	return found, sc.Err()
}

// redactSnippet keeps nearby context but replaces the matched secret so
// agent JSON/text never echoes the full credential (komando-style).
func redactSnippet(line string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(line) {
		end = len(line)
	}
	if start > end {
		start, end = end, start
	}
	const maxCtx = 24
	left := line[:start]
	if len(left) > maxCtx {
		left = "…" + left[len(left)-maxCtx:]
	}
	right := line[end:]
	if len(right) > maxCtx {
		right = right[:maxCtx] + "…"
	}
	return out.SafeText(strings.TrimSpace(left + "[redacted]" + right))
}
