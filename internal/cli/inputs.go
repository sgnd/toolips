package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
