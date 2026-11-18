// Package cli wires every toolips command.
package cli

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/sgnd/toolips/internal/out"
)

// Version is set at link time (-X) or from build info.
var Version = "dev"

type globalFlags struct {
	output string
	pretty bool
	dryRun bool
}

var g globalFlags

type exitCode int

func (e exitCode) Error() string { return "exit " + strconv.Itoa(int(e)) }

// Execute runs the CLI and returns a process exit code.
func Execute() int {
	return executeArgs(os.Args[1:])
}

func executeArgs(args []string) int {
	if Version == "dev" {
		if bi, ok := debug.ReadBuildInfo(); ok && isRelease(bi.Main.Version) {
			Version = bi.Main.Version
		}
	}
	out.P.Version = Version
	out.P.Start = time.Now()
	if f, ok := outputFormat(os.Getenv("TOOLIPS_OUTPUT")); ok {
		out.P.Format = f
	}
	root := newRoot()
	root.SetArgs(args)
	cmd, err := root.ExecuteC()
	if err == nil {
		return 0
	}
	var ec exitCode
	if errors.As(err, &ec) {
		return int(ec)
	}
	name := "toolips"
	if cmd != nil {
		name = commandID(cmd)
	}
	if f, ok := outputFormat(outputFlag()); ok {
		out.P.Format = f
	}
	var oe *out.Error
	if !errors.As(err, &oe) && isUsageErr(err) {
		err = out.Usage("%s", err.Error())
	}
	return out.P.Fail(name, err)
}

func isRelease(v string) bool {
	return strings.HasPrefix(v, "v") && !strings.Contains(v, "-")
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "toolips",
		Short: "All-in-one everyday tools for people and agents",
		Long: `All-in-one everyday tools for people and agents.

Private: pdf, image and secret commands work on local files only and never
use the network. Only tls, headers and ct connect to the host you name.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			out.P.Reset()
			out.P.Version = Version
			out.P.DryRun = g.dryRun
			if g.output == "" {
				g.output = os.Getenv("TOOLIPS_OUTPUT")
			}
			f, ok := outputFormat(g.output)
			if !ok {
				if g.output == "" || g.output == "auto" {
					f = autoFormat()
				} else {
					return out.Usage("-o must be text, json or auto (got %q)", g.output)
				}
			}
			out.P.Format = f
			out.P.Pretty = g.pretty || (f == "json" && term.IsTerminal(int(os.Stdout.Fd())))
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.PersistentFlags().StringVarP(&g.output, "output", "o", "auto", "output format: text, json or auto")
	root.PersistentFlags().BoolVar(&g.pretty, "pretty", false, "pretty-print JSON")
	root.PersistentFlags().BoolVar(&g.dryRun, "dry-run", false, "show what would change without writing")

	root.AddCommand(cmdVersion(), cmdList(), cmdGuide())
	root.AddCommand(cmdPDF(), cmdImage(), cmdTLS(), cmdHeaders(), cmdSecret(), cmdCT())
	root.InitDefaultCompletionCmd()
	markUsageErrors(root)
	return root
}

// markUsageErrors makes every argument and flag mistake a USAGE error (exit 2)
// by type, so it does not depend on cobra's message text.
func markUsageErrors(cmd *cobra.Command) {
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return out.Usage("%s", err.Error())
	})
	if args := cmd.Args; args != nil {
		cmd.Args = func(c *cobra.Command, a []string) error {
			if err := args(c, a); err != nil {
				return out.Usage("%s", err.Error())
			}
			return nil
		}
	}
	if cmd.HasSubCommands() && cmd.Run == nil && cmd.RunE == nil {
		// A group such as `pdf`: no args shows help, an unknown verb is USAGE.
		cmd.Args = cobra.ArbitraryArgs
		cmd.RunE = func(c *cobra.Command, a []string) error {
			if len(a) > 0 {
				return out.Usage("unknown command %q for %q", a[0], c.CommandPath())
			}
			return c.Help()
		}
	}
	for _, sub := range cmd.Commands() {
		markUsageErrors(sub)
	}
}

func autoFormat() string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "text"
	}
	return "json"
}

func outputFormat(s string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return "json", true
	case "text":
		return "text", true
	default:
		return "", false
	}
}

func outputFlag() string {
	args := os.Args[1:]
	for i := 0; i < len(args) && args[i] != "--"; i++ {
		a := args[i]
		switch {
		case (a == "-o" || a == "--output") && i+1 < len(args):
			return args[i+1]
		case strings.HasPrefix(a, "--output="):
			return strings.TrimPrefix(a, "--output=")
		case strings.HasPrefix(a, "-o") && len(a) > 2:
			return strings.TrimPrefix(a[2:], "=")
		}
	}
	if v := os.Getenv("TOOLIPS_OUTPUT"); v != "" {
		return v
	}
	return "auto"
}

func isUsageErr(err error) bool {
	s := err.Error()
	return strings.Contains(s, "unknown command") ||
		strings.Contains(s, "unknown flag") ||
		strings.Contains(s, "accepts") ||
		strings.Contains(s, "required flag") ||
		strings.Contains(s, "invalid argument")
}

func commandID(cmd *cobra.Command) string {
	parts := strings.Split(cmd.CommandPath(), " ")
	if len(parts) <= 1 {
		return "toolips"
	}
	return strings.Join(parts[1:], ".")
}

func done(cmd *cobra.Command, data any) error {
	out.P.Print(commandID(cmd), data, nil)
	return exitCode(0)
}

func writeFlag(cmd *cobra.Command) (outPath, outDir string) {
	outPath, _ = cmd.Flags().GetString("out")
	outDir, _ = cmd.Flags().GetString("outdir")
	return outPath, outDir
}

func addWriteFlags(cmd *cobra.Command) {
	cmd.Flags().String("out", "", "output file path")
	cmd.Flags().String("outdir", "", "output directory for batch writes")
}

func fileResultText(paths []string) string {
	if len(paths) == 0 {
		return "no files written\n"
	}
	var b strings.Builder
	for _, p := range paths {
		fmt.Fprintf(&b, "%s\n", out.SafeText(p))
	}
	return b.String()
}

type filesOut struct {
	Files []string `json:"files"`
	Count int      `json:"count"`
	Note  string   `json:"note,omitempty"`
}

func (f filesOut) Text() string {
	s := fileResultText(f.Files)
	if f.Note != "" {
		return s + f.Note + "\n"
	}
	return s
}

func newFilesOut(paths []string, note string) filesOut {
	return filesOut{Files: paths, Count: len(paths), Note: note}
}

// requireFiles returns a NOT_FOUND error (exit 3) for the first path that does
// not exist, so a missing input fails the same way in every command.
func requireFiles(paths []string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			return out.Errf(out.ExitNotFound, "NOT_FOUND", "check the file path", "%s: no such file", p)
		}
	}
	return nil
}
