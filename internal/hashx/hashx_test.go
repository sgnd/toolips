package hashx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sgnd/toolips/internal/hashx"
)

func TestSumKnownValues(t *testing.T) {
	p := filepath.Join(t.TempDir(), "abc.txt")
	if err := os.WriteFile(p, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"sha1":   "a9993e364706816aba3e25717850c26c9cd0d89d",
		"md5":    "900150983cd24fb0d6963f7d28e17f72",
	}
	for algo, sum := range want {
		got, err := hashx.Sum(p, algo)
		if err != nil || got != sum {
			t.Errorf("%s = %s, %v; want %s", algo, got, err, sum)
		}
	}
	if s, _ := hashx.Sum(p, "sha512"); len(s) != 128 {
		t.Errorf("sha512 length %d", len(s))
	}
	if _, err := hashx.Sum(p, "crc32"); err == nil {
		t.Error("unknown algorithm accepted")
	}
}

func TestVerify(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("abc"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.txt"), []byte("changed"), 0o644)
	sums := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad  a.txt\n" +
		"BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD *b.txt\n" +
		"\n" +
		"900150983cd24fb0d6963f7d28e17f72  gone.txt\n"
	sumFile := filepath.Join(dir, "checksums.txt")
	_ = os.WriteFile(sumFile, []byte(sums), 0o644)
	res, err := hashx.Verify(sumFile)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range res {
		got = append(got, r.File+"="+r.Status+"/"+r.Algo)
	}
	want := "a.txt=ok/sha256 b.txt=mismatch/sha256 gone.txt=missing/md5"
	if strings.Join(got, " ") != want {
		t.Errorf("got %v, want %s", got, want)
	}
}

func TestParseSumFileRejectsJunk(t *testing.T) {
	for _, in := range []string{"not a sum line\n", "zz  file\n", "abcd\n"} {
		if _, err := hashx.ParseSumFile(strings.NewReader(in)); err == nil {
			t.Errorf("accepted %q", in)
		}
	}
}
