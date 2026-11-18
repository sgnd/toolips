// Package hashx computes file checksums and checks sha256sum-style lists.
package hashx

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Algos are the supported algorithms. sha1 and md5 are for checking old
// sums only: they are broken for security use.
var Algos = []string{"sha256", "sha512", "sha1", "md5"}

func newHash(algo string) (hash.Hash, error) {
	switch algo {
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
	default:
		return nil, fmt.Errorf("algorithm must be one of %s (got %q)", strings.Join(Algos, ", "), algo)
	}
}

// Sum returns the hex checksum of the file at path, read as a stream.
func Sum(path, algo string) (string, error) {
	h, err := newHash(algo)
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Entry is one line of a sum file.
type Entry struct {
	Sum  string
	File string
}

// ParseSumFile reads "<hex>  <name>" lines (the sha256sum format; a "*"
// before the name, for binary mode, is allowed). Blank lines are skipped.
func ParseSumFile(r io.Reader) ([]Entry, error) {
	var out []Entry
	sc := bufio.NewScanner(r)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		sum, name, ok := strings.Cut(line, " ")
		name = strings.TrimPrefix(strings.TrimPrefix(name, " "), "*")
		if !ok || name == "" || !isHex(sum) {
			return nil, fmt.Errorf("line %d is not \"<checksum>  <file>\"", n)
		}
		out = append(out, Entry{Sum: strings.ToLower(sum), File: name})
	}
	return out, sc.Err()
}

// AlgoForLength picks the algorithm from a hex checksum's length.
func AlgoForLength(n int) (string, error) {
	switch n {
	case 64:
		return "sha256", nil
	case 128:
		return "sha512", nil
	case 40:
		return "sha1", nil
	case 32:
		return "md5", nil
	default:
		return "", fmt.Errorf("a %d-character checksum matches no supported algorithm", n)
	}
}

// Status of one file in Verify.
const (
	StatusOK       = "ok"
	StatusMismatch = "mismatch"
	StatusMissing  = "missing"
)

// Result is the outcome for one file in Verify.
type Result struct {
	File   string `json:"file"`
	Algo   string `json:"algo"`
	Status string `json:"status"`
}

// Verify checks every entry of sumFile. File names are relative to the
// folder that holds sumFile, as sha256sum writes them.
func Verify(sumFile string) ([]Result, error) {
	f, err := os.Open(sumFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entries, err := ParseSumFile(f)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(sumFile)
	results := make([]Result, 0, len(entries))
	for _, e := range entries {
		algo, err := AlgoForLength(len(e.Sum))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.File, err)
		}
		r := Result{File: e.File, Algo: algo}
		got, err := Sum(filepath.Join(dir, e.File), algo)
		switch {
		case err != nil:
			r.Status = StatusMissing
		case got == e.Sum:
			r.Status = StatusOK
		default:
			r.Status = StatusMismatch
		}
		results = append(results, r)
	}
	return results, nil
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
