package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sgnd/toolips/internal/out"
)

// expandInputs turns file arguments, quoted globs and folders into a sorted
// list of paths that keep accepts. Folders are read one level deep. Agents
// often pass a literal "*.jpg"; shells may already expand it, so plain files
// are kept as they are. kind names the files in errors ("images", "files").
func expandInputs(args []string, keep func(string) bool, kind string) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string
	add := func(paths ...string) {
		for _, p := range paths {
			if _, ok := seen[p]; ok || p == "" {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	for _, arg := range args {
		if strings.ContainsAny(arg, "*?[") {
			matches, err := filepath.Glob(arg)
			if err != nil {
				return nil, err
			}
			var kept []string
			for _, m := range matches {
				if st, err := os.Stat(m); err == nil && !st.IsDir() && keep(m) {
					kept = append(kept, m)
				}
			}
			if len(kept) == 0 {
				return nil, fmt.Errorf("no %s matched %q", kind, arg)
			}
			sort.Strings(kept)
			add(kept...)
			continue
		}
		st, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		if !st.IsDir() {
			add(arg)
			continue
		}
		entries, err := os.ReadDir(arg)
		if err != nil {
			return nil, err
		}
		var kept []string
		for _, e := range entries {
			p := filepath.Join(arg, e.Name())
			if !e.IsDir() && keep(p) {
				kept = append(kept, p)
			}
		}
		if len(kept) == 0 {
			return nil, fmt.Errorf("no %s in folder %q", kind, arg)
		}
		sort.Strings(kept)
		add(kept...)
	}
	if len(out) == 0 {
		return nil, os.ErrNotExist
	}
	return out, nil
}

func anyFile(string) bool { return true }

// readSecretLine reads the first line of path, or of stdin when path is "-".
// Secrets (passwords, tokens) come this way, never as an argument, so they
// do not show in the process list or shell history. what names the secret
// in errors; hint says where to look; limit caps the bytes read.
func readSecretLine(cmd *cobra.Command, path, what, hint string, limit int64) (string, error) {
	var r io.Reader = cmd.InOrStdin()
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return "", out.Errf(out.ExitNotFound, "NOT_FOUND", hint, "%s: no such file", path)
		}
		defer f.Close()
		r = f
	}
	line, err := bufio.NewReader(io.LimitReader(r, limit)).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", out.Errf(out.ExitError, "READ", hint, "read %s: %v", what, err)
	}
	v := strings.TrimRight(line, "\r\n")
	if v == "" {
		return "", out.Usage("the %s is empty", what)
	}
	return v, nil
}
