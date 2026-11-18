package pdfx

import (
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestConfIsOffline(t *testing.T) {
	if !conf().Offline {
		t.Fatal("pdfcpu must run offline so PDFs never leave the machine")
	}
}

func TestPDFDate(t *testing.T) {
	cases := map[string]string{
		"D:20261118111924+07'00'": "2026-11-18T11:19:24+07:00",
		"D:20261118111924Z":       "2026-11-18T11:19:24Z",
		"D:20261118111924":        "2026-11-18T11:19:24Z",
		"not a date":              "not a date",
		"":                        "",
	}
	for in, want := range cases {
		if got := pdfDate(in); got != want {
			t.Errorf("pdfDate(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestProtectUnlock(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "p.png")
	f, err := os.Create(img)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, image.NewGray(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	ctx := context.Background()
	src := filepath.Join(dir, "a.pdf")
	if err := FromImage(ctx, []string{img}, src); err != nil {
		t.Fatal(err)
	}
	locked := filepath.Join(dir, "locked.pdf")
	if err := Protect(ctx, src, locked, "s3cret-pass"); err != nil {
		t.Fatal(err)
	}
	if info, err := ReadInfo(ctx, locked); err == nil && !info.Encrypted {
		t.Error("protected file is not encrypted")
	}
	if err := Unlock(ctx, locked, filepath.Join(dir, "bad.pdf"), "wrong"); err == nil {
		t.Error("wrong password unlocked the file")
	}
	open := filepath.Join(dir, "open.pdf")
	if err := Unlock(ctx, locked, open, "s3cret-pass"); err != nil {
		t.Fatal(err)
	}
	if info, err := ReadInfo(ctx, open); err != nil || info.Encrypted {
		t.Errorf("unlocked info = %+v, %v", info, err)
	}
	if err := Protect(ctx, src, locked, ""); err == nil {
		t.Error("empty password accepted")
	}
}
