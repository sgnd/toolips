package imgx

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// exifWith builds an APP1 EXIF segment with orientation o and a fake
// "GPS-SECRET" marker standing in for location data.
func exifWith(o int) []byte {
	seg := orientationSegment(o)
	seg = append(seg, []byte("GPS-SECRET")...)
	binary.BigEndian.PutUint16(seg[2:], uint16(len(seg)-2))
	return seg
}

func TestStripJPEGKeepsOrientation(t *testing.T) {
	var enc bytes.Buffer
	if err := jpeg.Encode(&enc, image.NewRGBA(image.Rect(0, 0, 8, 4)), nil); err != nil {
		t.Fatal(err)
	}
	raw := enc.Bytes()
	com := []byte{0xFF, 0xFE, 0, 9, 'c', 'a', 'm', 'e', 'r', 'a', '1'}
	var src []byte
	src = append(src, raw[:2]...)
	src = append(src, exifWith(6)...)
	src = append(src, com...)
	src = append(src, raw[2:]...)

	dir := t.TempDir()
	in, out := filepath.Join(dir, "a.jpg"), filepath.Join(dir, "b.jpg")
	if err := os.WriteFile(in, src, 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Strip(in, out)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(out)
	if bytes.Contains(got, []byte("GPS-SECRET")) || bytes.Contains(got, []byte("camera1")) {
		t.Error("metadata left in the output")
	}
	if res.Orientation != 6 || len(res.Removed) != 2 {
		t.Errorf("result = %+v", res)
	}
	if _, err := jpeg.Decode(bytes.NewReader(got)); err != nil {
		t.Errorf("output does not decode: %v", err)
	}
	i := bytes.Index(got, []byte("Exif\x00\x00"))
	if i < 0 || exifOrientation(got[i+6:]) != 6 {
		t.Error("orientation 6 not kept")
	}
}

func TestStripPNGDropsText(t *testing.T) {
	var enc bytes.Buffer
	if err := png.Encode(&enc, image.NewGray(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	raw := enc.Bytes()
	body := []byte("Comment\x00GPS-SECRET")
	chunk := make([]byte, 8, 12+len(body))
	binary.BigEndian.PutUint32(chunk, uint32(len(body)))
	copy(chunk[4:], "tEXt")
	chunk = append(chunk, body...)
	chunk = binary.BigEndian.AppendUint32(chunk, crc32.ChecksumIEEE(chunk[4:]))
	ihdrEnd := 8 + 25
	src := append(append(append([]byte{}, raw[:ihdrEnd]...), chunk...), raw[ihdrEnd:]...)

	dir := t.TempDir()
	in, out := filepath.Join(dir, "a.png"), filepath.Join(dir, "b.png")
	_ = os.WriteFile(in, src, 0o644)
	res, err := Strip(in, out)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(out)
	if bytes.Contains(got, []byte("GPS-SECRET")) || len(res.Removed) != 1 || res.Removed[0] != "text" {
		t.Errorf("result = %+v", res)
	}
	if _, err := png.Decode(bytes.NewReader(got)); err != nil {
		t.Errorf("output does not decode: %v", err)
	}
}

func TestStripRefusesOtherFormats(t *testing.T) {
	var enc bytes.Buffer
	_ = gif.Encode(&enc, image.NewPaletted(image.Rect(0, 0, 2, 2), nil), nil)
	in := filepath.Join(t.TempDir(), "a.gif")
	_ = os.WriteFile(in, enc.Bytes(), 0o644)
	if _, err := Strip(in, in+".out"); err == nil {
		t.Error("GIF accepted")
	}
}
