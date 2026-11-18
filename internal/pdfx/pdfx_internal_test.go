package pdfx

import "testing"

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
