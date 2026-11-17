// Package out renders every command result as a stable, machine-readable
// envelope. AI agents parse stdout; humans can ask for -o text.
package out

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Exit codes are part of the public contract. Agents may branch on them.
const (
	ExitOK       = 0
	ExitError    = 1 // generic / internal error
	ExitUsage    = 2 // bad flags or arguments
	ExitNotFound = 3 // file or resource not found
	ExitNetwork  = 4 // network / TLS failure
)

// Contract is the version of toolips's output contract.
const Contract = 1

// UsageHint is the default hint for usage errors.
const UsageHint = "run with --help or `toolips list` for available commands"

// Error is a typed error carrying a stable code, an exit code and a hint.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Exit    int    `json:"-"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func (e *Error) WithHint(h string) *Error { e.Hint = h; return e }

func (e Error) MarshalJSON() ([]byte, error) {
	type plain Error
	p := plain(e)
	if p.Hint == "" {
		p.Hint = defaultHint(e.Exit)
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(p); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b.Bytes(), []byte("\n")), nil
}

func defaultHint(exit int) string {
	switch exit {
	case ExitUsage:
		return UsageHint
	case ExitNotFound:
		return "check the path; `toolips list` shows the commands"
	case ExitNetwork:
		return "check the host, port and network connection"
	}
	return "read the message; run with --help for usage"
}

// Errf builds a typed error.
func Errf(exit int, code, hint, format string, a ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, a...), Hint: hint, Exit: exit}
}

// Usage is a usage/argument error (exit 2).
func Usage(format string, a ...any) *Error {
	return Errf(ExitUsage, "USAGE", UsageHint, format, a...)
}

// Envelope is the single JSON object printed on stdout for every command.
type Envelope struct {
	OK      bool   `json:"ok"`
	Command string `json:"command"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    Meta   `json:"meta"`
}

// Meta holds contract metadata for every envelope.
type Meta struct {
	Contract   int    `json:"contract"`
	Version    string `json:"version"`
	DurationMS int64  `json:"duration_ms"`
	DryRun     bool   `json:"dry_run,omitempty"`
}

// Printer is configured once by the root command.
type Printer struct {
	Format  string // json | text
	Pretty  bool
	Version string
	Start   time.Time
	DryRun  bool
	W       io.Writer
}

// P is the process-wide printer.
var P = &Printer{Format: "json", W: os.Stdout, Start: time.Now()}

// Texter lets a result type provide a human rendering for -o text.
type Texter interface{ Text() string }

func (p *Printer) meta() Meta {
	return Meta{
		Contract:   Contract,
		Version:    p.Version,
		DurationMS: time.Since(p.Start).Milliseconds(),
		DryRun:     p.DryRun,
	}
}

// Reset forgets what one command recorded (tests run many in one process).
func (p *Printer) Reset() {
	p.DryRun = false
	p.Start = time.Now()
}

// Print prints data. A non-nil softErr marks the envelope not-ok but still
// carries data.
func (p *Printer) Print(command string, data any, softErr *Error) {
	env := Envelope{OK: softErr == nil, Command: command, Data: data, Error: softErr, Meta: p.meta()}
	if p.Format == "text" {
		if p.DryRun {
			fmt.Fprintln(os.Stderr, "Dry run: nothing was changed")
		}
		if t, ok := data.(Texter); ok {
			s := t.Text()
			fmt.Fprint(p.W, s)
			if s != "" && !strings.HasSuffix(s, "\n") {
				fmt.Fprintln(p.W)
			}
		} else if data != nil {
			s := renderText(data)
			fmt.Fprint(p.W, s)
			if s == "" {
				fmt.Fprintln(os.Stderr, "Nothing to show")
			}
		}
		if softErr != nil {
			p.printErr(softErr)
		}
		return
	}
	p.writeJSON(env)
}

// Fail prints an error envelope (or text) and returns the exit code.
func (p *Printer) Fail(command string, err error) int {
	var e *Error
	if !errors.As(err, &e) {
		e = &Error{Code: "ERROR", Message: err.Error(), Exit: ExitError}
	}
	if e.Hint == "" {
		e.Hint = defaultHint(e.Exit)
	}
	if p.Format == "text" {
		p.printErr(e)
		return e.Exit
	}
	p.writeJSON(Envelope{OK: false, Command: command, Error: e, Meta: p.meta()})
	return e.Exit
}

func (p *Printer) printErr(e *Error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", e.Message)
	if e.Hint != "" {
		fmt.Fprintf(os.Stderr, "Hint: %s\n", e.Hint)
	}
	fmt.Fprintf(os.Stderr, "(code %s, exit %d)\n", e.Code, e.Exit)
}

func (p *Printer) writeJSON(v any) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if p.Pretty {
		enc.SetIndent("", "  ")
	}
	_ = enc.Encode(v)
	_, _ = p.W.Write(b.Bytes())
}

func renderText(data any) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprint(data)
	}
	return string(b) + "\n"
}
