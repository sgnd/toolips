package pdfx

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Info is a short summary of one PDF.
type Info struct {
	File       string `json:"file"`
	SizeBytes  int64  `json:"size_bytes"`
	Version    string `json:"version"`
	Pages      int    `json:"pages"`
	PageSize   string `json:"page_size,omitempty"` // first page, in points
	Title      string `json:"title,omitempty"`
	Author     string `json:"author,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Creator    string `json:"creator,omitempty"`
	Producer   string `json:"producer,omitempty"`
	Created    string `json:"created,omitempty"`
	Modified   string `json:"modified,omitempty"`
	Encrypted  bool   `json:"encrypted"`
	Form       bool   `json:"form"`
	Signatures bool   `json:"signatures"`
}

// ReadInfo reads the summary of inFile without changing it.
func ReadInfo(ctx context.Context, inFile string) (Info, error) {
	f, err := os.Open(inFile)
	if err != nil {
		return Info{}, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return Info{}, err
	}
	pi, err := api.PDFInfo(ctx, f, inFile, nil, false, conf())
	if err != nil {
		return Info{}, err
	}
	info := Info{
		File: inFile, SizeBytes: st.Size(), Version: pi.Version, Pages: pi.PageCount,
		Title: pi.Title, Author: pi.Author, Subject: pi.Subject,
		Creator: pi.Creator, Producer: pi.Producer,
		Created: pdfDate(pi.CreationDate), Modified: pdfDate(pi.ModificationDate),
		Encrypted: pi.Encrypted, Form: pi.Form, Signatures: pi.Signatures,
	}
	if _, err := f.Seek(0, io.SeekStart); err == nil {
		if dims, err := api.PageDims(ctx, f, conf()); err == nil && len(dims) > 0 {
			info.PageSize = fmt.Sprintf("%.0fx%.0f", dims[0].Width, dims[0].Height)
		}
	}
	return info, nil
}

// pdfDate turns a PDF date ("D:20261118111924+07'00'") into RFC 3339, or
// returns it unchanged when it does not parse.
func pdfDate(s string) string {
	v := strings.TrimPrefix(s, "D:")
	v = strings.ReplaceAll(v, "'", "")
	for _, layout := range []string{"20060102150405-0700", "20060102150405Z", "20060102150405"} {
		if t, err := time.Parse(layout, v); err == nil {
			return t.Format(time.RFC3339)
		}
	}
	return s
}
