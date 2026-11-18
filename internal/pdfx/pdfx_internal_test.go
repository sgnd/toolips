package pdfx

import "testing"

func TestConfIsOffline(t *testing.T) {
	if !conf().Offline {
		t.Fatal("pdfcpu must run offline so PDFs never leave the machine")
	}
}
