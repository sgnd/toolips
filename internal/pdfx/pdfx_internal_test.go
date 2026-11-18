package pdfx

import (
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
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

func TestNumberPages(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "p.png")
	f, err := os.Create(img)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, image.NewGray(image.Rect(0, 0, 200, 200))); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	ctx := context.Background()
	src := filepath.Join(dir, "a.pdf")
	if err := FromImage(ctx, []string{img, img}, src); err != nil {
		t.Fatal(err)
	}
	numbered := filepath.Join(dir, "n.pdf")
	if err := NumberPages(ctx, src, numbered, "%p of %P", "bc", nil); err != nil {
		t.Fatal(err)
	}
	r, err := os.Open(numbered)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	pi, err := api.PDFInfo(ctx, r, numbered, nil, false, conf())
	if err != nil {
		t.Fatal(err)
	}
	if !pi.Watermarked || pi.PageCount != 2 {
		t.Errorf("numbered: watermarked=%v pages=%d", pi.Watermarked, pi.PageCount)
	}
	if err := NumberPages(ctx, src, filepath.Join(dir, "x.pdf"), "", "middle", nil); err == nil {
		t.Error("bad position accepted")
	}
}

func bombPDF(t testing.TB, inflated int) string {
	var z bytes.Buffer
	w, _ := zlib.NewWriterLevel(&z, zlib.BestCompression)
	_, _ = w.Write(make([]byte, inflated))
	_ = w.Close()
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 100 100] /Contents 4 0 R >>",
		fmt.Sprintf("<< /Length %d /Filter /FlateDecode >>\nstream\n%s\nendstream", z.Len(), z.Bytes()),
	}
	var b bytes.Buffer
	b.WriteString("%PDF-1.7\n")
	offs := []int{}
	for i, o := range objs {
		offs = append(offs, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	x := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, o := range offs {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, x)
	p := filepath.Join(t.TempDir(), "bomb.pdf")
	_ = os.WriteFile(p, b.Bytes(), 0o644)
	return p
}

// TestBombIsRefused: a 4 MiB zero stream against a 1 MiB limit. Every
// command that decodes page content must stop with the limit error.
func TestBombIsRefused(t *testing.T) {
	old := Limits
	Limits.DecodeBytes = 1 << 20
	defer func() { Limits = old }()
	p := bombPDF(t, 4<<20)
	ctx := context.Background()
	d := t.TempDir()
	errs := map[string]error{
		"compress":  Compress(ctx, p, filepath.Join(d, "c.pdf")),
		"number":    NumberPages(ctx, p, filepath.Join(d, "n.pdf"), "%p", "bc", nil),
		"watermark": Watermark(ctx, p, filepath.Join(d, "w.pdf"), "x", "", nil),
		"pages":     KeepPages(ctx, p, filepath.Join(d, "k.pdf"), "1"),
	}
	for name, err := range errs {
		if !IsTooLarge(err) {
			t.Errorf("%s: got %v, want the decode limit error", name, err)
		}
	}
}

func TestInputSizeLimit(t *testing.T) {
	old := Limits
	Limits.InputBytes = 100
	defer func() { Limits = old }()
	p := bombPDF(t, 10)
	if _, err := ReadInfo(context.Background(), p); !IsTooLarge(err) {
		t.Errorf("got %v, want the input size limit error", err)
	}
}

// FuzzReadInfo and FuzzCompress feed arbitrary bytes to the PDF reader.
// Run them with a time and memory limit, as docs/security records.
func FuzzReadInfo(f *testing.F) {
	fuzzSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		p := filepath.Join(t.TempDir(), "f.pdf")
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
		withinTime(t, func() { _, _ = ReadInfo(context.Background(), p) })
	})
}

func FuzzCompress(f *testing.F) {
	fuzzSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		d := t.TempDir()
		p := filepath.Join(d, "f.pdf")
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
		withinTime(t, func() { _ = Compress(context.Background(), p, filepath.Join(d, "out.pdf")) })
	})
}

func fuzzSeeds(f *testing.F) {
	for _, n := range []int{0, 100, 64 << 10} {
		b, err := os.ReadFile(bombPDF(f, n))
		if err == nil {
			f.Add(b)
		}
	}
	f.Add([]byte("%PDF-1.7\n%%EOF\n"))
}

// withinTime fails the fuzz input when fn runs longer than 5 seconds, so
// the fuzzer saves inputs that hang, not only inputs that crash.
func withinTime(t *testing.T, fn func()) {
	done := make(chan struct{})
	go func() { fn(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("took longer than 5s")
	}
}

func TestNUpAndBooklet(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "p.png")
	f, _ := os.Create(img)
	_ = png.Encode(f, image.NewGray(image.Rect(0, 0, 50, 70)))
	_ = f.Close()
	ctx := context.Background()
	src := filepath.Join(dir, "a.pdf")
	if err := FromImage(ctx, []string{img, img, img, img}, src); err != nil {
		t.Fatal(err)
	}
	for name, fn := range map[string]func(string) error{
		"nup":     func(o string) error { return NUp(ctx, src, o, 4, "A4") },
		"booklet": func(o string) error { return Booklet(ctx, src, o, 2, "A4") },
	} {
		o := filepath.Join(dir, name+".pdf")
		if err := fn(o); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		info, err := ReadInfo(ctx, o)
		if err != nil || info.Pages < 1 || info.Pages >= 4 {
			t.Errorf("%s: %d pages, %v; want fewer sheets than the 4 pages", name, info.Pages, err)
		}
	}
	if err := NUp(ctx, src, filepath.Join(dir, "x.pdf"), 5, "A4"); err == nil {
		t.Error("5 per sheet accepted")
	}
}
