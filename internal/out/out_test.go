package out

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Pretty: true, Version: "test", Start: time.Now(), W: &buf}
	p.Print("version", map[string]string{"version": "test"}, nil)
	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK || env.Command != "version" || env.Meta.Contract != Contract {
		t.Fatalf("bad envelope: %+v", env)
	}
}

func TestFailUsage(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "json", Version: "test", Start: time.Now(), W: &buf}
	code := p.Fail("pdf.merge", Usage("need at least two files"))
	if code != ExitUsage {
		t.Fatalf("exit=%d", code)
	}
	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.OK || env.Error == nil || env.Error.Code != "USAGE" {
		t.Fatalf("bad fail: %+v", env)
	}
}

func TestPrintText(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "text", Version: "test", Start: time.Now(), W: &buf}
	p.Print("list", textResult{s: "pdf\nimage\n"}, nil)
	if !strings.Contains(buf.String(), "pdf") {
		t.Fatalf("got %q", buf.String())
	}
}

type textResult struct{ s string }

func (t textResult) Text() string { return t.s }

type texter struct{ s string }

func (t texter) Text() string { return t.s }

func TestPrintTextCases(t *testing.T) {
	cases := []struct {
		name string
		data any
		want string
	}{
		{"texter adds newline", texter{"hi"}, "hi\n"},
		{"texter keeps newline", texter{"hi\n"}, "hi\n"},
		{"plain data as JSON", map[string]int{"n": 1}, "{\n  \"n\": 1\n}\n"},
		{"nil prints nothing", nil, ""},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		p := &Printer{Format: "text", W: &buf, Start: time.Now(), DryRun: true}
		p.Print("x", c.data, nil)
		if buf.String() != c.want {
			t.Errorf("%s: got %q, want %q", c.name, buf.String(), c.want)
		}
	}
}

func TestFailTextAndDefaults(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "text", W: &buf, Start: time.Now()}
	if code := p.Fail("x", errors.New("boom")); code != ExitError {
		t.Fatalf("exit %d", code)
	}
	if buf.Len() != 0 {
		t.Fatalf("text errors go to stderr, stdout got %q", buf.String())
	}
	p.Print("x", texter{"partial"}, Errf(ExitNetwork, "NET", "", "down"))

	for _, exit := range []int{ExitError, ExitUsage, ExitNotFound, ExitNetwork} {
		if defaultHint(exit) == "" {
			t.Errorf("no default hint for exit %d", exit)
		}
	}
	e := Errf(ExitError, "CODE", "", "msg").WithHint("do this")
	if e.Hint != "do this" || !strings.Contains(e.Error(), "CODE: msg") {
		t.Errorf("got %+v / %q", e, e.Error())
	}
	p.DryRun = true
	p.Reset()
	if p.DryRun {
		t.Error("Reset kept DryRun")
	}
}
