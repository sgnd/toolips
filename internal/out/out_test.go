package out

import (
	"bytes"
	"encoding/json"
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
