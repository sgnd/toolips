package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sgnd/toolips/internal/hashx"
	"github.com/sgnd/toolips/internal/out"
)

func cmdHash() *cobra.Command {
	h := &cobra.Command{Use: "hash", Short: "File checksums"}
	h.AddCommand(hashSum(), hashVerify())
	return h
}

func hashSum() *cobra.Command {
	var algo string
	cmd := &cobra.Command{
		Use:   "sum FILE [FILE...]",
		Short: "Print the checksum of files (SHA-256 by default)",
		Long: `Print the checksum of files (SHA-256 by default).

Takes several files, a folder (one level deep) or a quoted glob. Text output
is in the sha256sum format, so it can be saved and checked later with
"toolips hash verify". sha1 and md5 are for checking old sums only; they are
not safe for security use.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !slices.Contains(hashx.Algos, algo) {
				return out.Usage("--algo must be one of %s", strings.Join(hashx.Algos, ", "))
			}
			files, err := expandInputs(args, anyFile, "files")
			if err != nil {
				return out.Errf(out.ExitNotFound, "NOT_FOUND", "check the file paths", "%v", err)
			}
			res := sumOut{Algo: algo, Files: []sumLine{}}
			for _, f := range files {
				s, err := hashx.Sum(f, algo)
				if err != nil {
					return out.Errf(out.ExitError, "HASH", "check that the file can be read", "%s: %v", f, err)
				}
				res.Files = append(res.Files, sumLine{File: f, Sum: s})
			}
			return done(cmd, res)
		},
	}
	cmd.Flags().StringVar(&algo, "algo", "sha256", "sha256, sha512, sha1 or md5")
	return cmd
}

type sumLine struct {
	File string `json:"file"`
	Sum  string `json:"sum"`
}

type sumOut struct {
	Algo  string    `json:"algo"`
	Files []sumLine `json:"files"`
}

func (s sumOut) Text() string {
	var b strings.Builder
	for _, f := range s.Files {
		fmt.Fprintf(&b, "%s  %s\n", f.Sum, out.SafeText(f.File))
	}
	return b.String()
}

func hashVerify() *cobra.Command {
	return &cobra.Command{
		Use:   "verify SUMFILE",
		Short: "Check files against a sha256sum-style list, such as a release's checksums.txt",
		Long: `Check files against a sha256sum-style list, such as a release's checksums.txt.

Each line is "<checksum>  <file>", with file names relative to the list's
folder. The algorithm follows from the checksum length (SHA-256, SHA-512,
SHA-1 or MD5). Every file is reported as ok, mismatch or missing; the
command fails (HASH_MISMATCH, exit 1) when any file is not ok.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			results, err := hashx.Verify(args[0])
			if err != nil {
				return out.Errf(out.ExitError, "HASH", "check that the list is in the sha256sum format", "%v", err)
			}
			res := verifyOut{Files: results}
			for _, r := range results {
				if r.Status == hashx.StatusOK {
					res.OK++
				} else {
					res.Failed++
				}
			}
			if res.Failed > 0 {
				e := out.Errf(out.ExitError, "HASH_MISMATCH", "a file changed or is missing; download it again", "%d of %d files did not match", res.Failed, len(results))
				out.P.Print(commandID(cmd), res, e)
				return exitCode(e.Exit)
			}
			return done(cmd, res)
		},
	}
}

type verifyOut struct {
	Files  []hashx.Result `json:"files"`
	OK     int            `json:"ok"`
	Failed int            `json:"failed"`
}

func (v verifyOut) Text() string {
	var b strings.Builder
	for _, f := range v.Files {
		fmt.Fprintf(&b, "%-8s  %s\n", f.Status, out.SafeText(f.File))
	}
	fmt.Fprintf(&b, "%d ok, %d failed\n", v.OK, v.Failed)
	return b.String()
}
