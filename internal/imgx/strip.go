package imgx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
)

// StripResult says what Strip removed from one file.
type StripResult struct {
	Removed     []string `json:"removed"`
	Orientation int      `json:"orientation,omitempty"` // kept so the photo stays upright
}

// Strip copies in to out without metadata, byte for byte otherwise, so the
// picture is not re-encoded. JPEG loses EXIF, XMP (APP1), IPTC (APP13) and
// comments, keeping only the EXIF orientation. PNG loses text, eXIf and time
// chunks. Other formats are refused, because they would need re-encoding.
func Strip(in, out string) (StripResult, error) {
	data, err := os.ReadFile(in)
	if err != nil {
		return StripResult{}, err
	}
	var res StripResult
	var clean []byte
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8}):
		clean, res, err = stripJPEG(data)
	case bytes.HasPrefix(data, pngMagic):
		clean, res, err = stripPNG(data)
	default:
		return res, fmt.Errorf("strip works on JPG and PNG only; convert %s first", filepath.Base(in))
	}
	if err != nil {
		return res, err
	}
	if res.Removed == nil {
		res.Removed = []string{}
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return res, err
	}
	return res, os.WriteFile(out, clean, 0o644)
}

var pngMagic = []byte("\x89PNG\r\n\x1a\n")

func stripJPEG(data []byte) ([]byte, StripResult, error) {
	var res StripResult
	var b bytes.Buffer
	b.Write(data[:2])
	i := 2
	for i+4 <= len(data) {
		if data[i] != 0xFF {
			return nil, res, errors.New("broken JPEG: marker expected")
		}
		marker := data[i+1]
		if marker == 0xDA { // start of scan: the rest is picture data
			b.Write(data[i:])
			return b.Bytes(), res, nil
		}
		n := int(binary.BigEndian.Uint16(data[i+2:]))
		end := i + 2 + n
		if n < 2 || end > len(data) {
			return nil, res, errors.New("broken JPEG: bad segment length")
		}
		seg := data[i:end]
		switch marker {
		case 0xE1: // EXIF or XMP
			payload := seg[4:]
			if bytes.HasPrefix(payload, []byte("Exif\x00\x00")) {
				res.Removed = append(res.Removed, "exif")
				if o := exifOrientation(payload[6:]); o > 1 {
					res.Orientation = o
					b.Write(orientationSegment(o))
				}
			} else {
				res.Removed = append(res.Removed, "xmp")
			}
		case 0xED:
			res.Removed = append(res.Removed, "iptc")
		case 0xFE:
			res.Removed = append(res.Removed, "comment")
		default:
			b.Write(seg)
		}
		i = end
	}
	return nil, res, errors.New("broken JPEG: no picture data")
}

// exifOrientation reads tag 0x0112 from IFD0 of a TIFF-format EXIF block.
func exifOrientation(t []byte) int {
	if len(t) < 8 {
		return 0
	}
	var bo binary.ByteOrder
	switch string(t[:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 0
	}
	off := int(bo.Uint32(t[4:]))
	if off+2 > len(t) {
		return 0
	}
	count := int(bo.Uint16(t[off:]))
	for k := 0; k < count; k++ {
		e := off + 2 + k*12
		if e+12 > len(t) {
			return 0
		}
		if bo.Uint16(t[e:]) == 0x0112 {
			if o := int(bo.Uint16(t[e+8:])); o >= 1 && o <= 8 {
				return o
			}
			return 0
		}
	}
	return 0
}

// orientationSegment is an APP1 EXIF segment holding only the orientation.
func orientationSegment(o int) []byte {
	tiff := []byte{'I', 'I', 42, 0, 8, 0, 0, 0, // header, IFD0 at 8
		1, 0, // one entry
		0x12, 0x01, 3, 0, 1, 0, 0, 0, byte(o), 0, 0, 0, // orientation, SHORT, 1
		0, 0, 0, 0} // no next IFD
	payload := append([]byte("Exif\x00\x00"), tiff...)
	seg := []byte{0xFF, 0xE1, 0, 0}
	binary.BigEndian.PutUint16(seg[2:], uint16(len(payload)+2))
	return append(seg, payload...)
}

var pngDrop = map[string]string{"tEXt": "text", "zTXt": "text", "iTXt": "text", "eXIf": "exif", "tIME": "time"}

func stripPNG(data []byte) ([]byte, StripResult, error) {
	var res StripResult
	var b bytes.Buffer
	b.Write(pngMagic)
	seen := map[string]bool{}
	i := len(pngMagic)
	for i+12 <= len(data) {
		n := int(binary.BigEndian.Uint32(data[i:]))
		end := i + 12 + n
		if n < 0 || end > len(data) {
			return nil, res, errors.New("broken PNG: bad chunk length")
		}
		typ := string(data[i+4 : i+8])
		chunk := data[i:end]
		if crc32.ChecksumIEEE(chunk[4:8+n]) != binary.BigEndian.Uint32(chunk[8+n:]) {
			return nil, res, fmt.Errorf("broken PNG: bad checksum in %s", typ)
		}
		if name, drop := pngDrop[typ]; drop {
			if !seen[name] {
				res.Removed = append(res.Removed, name)
				seen[name] = true
			}
		} else {
			b.Write(chunk)
		}
		i = end
		if typ == "IEND" {
			return b.Bytes(), res, nil
		}
	}
	return nil, res, errors.New("broken PNG: no end chunk")
}
