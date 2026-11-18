package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sgnd/toolips/internal/out"
	"github.com/sgnd/toolips/internal/pdfx"
)

func cmdPDF() *cobra.Command {
	pdf := &cobra.Command{Use: "pdf", Short: "PDF tools"}
	pdf.AddCommand(
		pdfMerge(), pdfSplit(), pdfCompress(), pdfRotate(),
		pdfWatermark(), pdfToImage(), pdfFromImage(), pdfInfo(),
		pdfProtect(), pdfUnlock(), pdfPages(), pdfNumber(),
	)
	return pdf
}

func pdfMerge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge FILE FILE [FILE...]",
		Short: "Merge PDF files into one",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = "merged.pdf"
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, "would merge "+strconv.Itoa(len(args))+" files"))
			}
			if err := pdfx.Merge(context.Background(), args, outPath); err != nil {
				return pdfErr("PDF_MERGE", "check the input PDFs", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

func pdfSplit() *cobra.Command {
	var span int
	cmd := &cobra.Command{
		Use:   "split FILE",
		Short: "Split a PDF into smaller files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			_, outDir := writeFlag(cmd)
			if outDir == "" {
				outDir = strings.TrimSuffix(args[0], filepath.Ext(args[0])) + "_pages"
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outDir}, "would split into "+outDir))
			}
			if err := pdfx.Split(context.Background(), args[0], outDir, span); err != nil {
				return pdfErr("PDF_SPLIT", "check the input PDF", err)
			}
			files, _ := pdfx.ListDirFiles(outDir)
			return done(cmd, newFilesOut(files, ""))
		},
	}
	cmd.Flags().IntVar(&span, "span", 1, "pages per output file")
	addWriteFlags(cmd)
	return cmd
}

func pdfCompress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compress FILE",
		Short: "Compress (optimize) a PDF",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = pdfx.DefaultOut(args[0], ".compressed", ".pdf")
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, ""))
			}
			if err := pdfx.Compress(context.Background(), args[0], outPath); err != nil {
				return pdfErr("PDF_COMPRESS", "check the input PDF", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

func pdfRotate() *cobra.Command {
	var degrees int
	var pages string
	cmd := &cobra.Command{
		Use:   "rotate FILE",
		Short: "Rotate PDF pages (90/180/270)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = pdfx.DefaultOut(args[0], ".rotated", ".pdf")
			}
			var sel []string
			if pages != "" {
				sel = []string{pages}
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, ""))
			}
			if err := pdfx.Rotate(context.Background(), args[0], outPath, degrees, sel); err != nil {
				return pdfErr("PDF_ROTATE", "use --degrees 90, 180 or 270", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	cmd.Flags().IntVar(&degrees, "degrees", 90, "clockwise rotation")
	cmd.Flags().StringVar(&pages, "pages", "", "page selection (pdfcpu syntax, empty=all)")
	addWriteFlags(cmd)
	return cmd
}

func pdfWatermark() *cobra.Command {
	var text, desc, pages string
	cmd := &cobra.Command{
		Use:   "watermark FILE",
		Short: "Add a text watermark to a PDF",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			if text == "" {
				return out.Usage("--text is required")
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = pdfx.DefaultOut(args[0], ".wm", ".pdf")
			}
			var sel []string
			if pages != "" {
				sel = []string{pages}
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, ""))
			}
			if err := pdfx.Watermark(context.Background(), args[0], outPath, text, desc, sel); err != nil {
				return pdfErr("PDF_WATERMARK", "check --text and the input PDF", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "watermark text")
	cmd.Flags().StringVar(&desc, "desc", "", "pdfcpu watermark description")
	cmd.Flags().StringVar(&pages, "pages", "", "page selection (empty=all)")
	addWriteFlags(cmd)
	return cmd
}

func pdfToImage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-image FILE",
		Short: "Extract embedded images from a PDF (not page rasterization)",
		Long: `Extract embedded images from a PDF into a directory.

This does not rasterize pages. True page-to-image rendering needs CGO
(and is intentionally omitted so toolips stays CGO_ENABLED=0).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			_, outDir := writeFlag(cmd)
			if outDir == "" {
				outDir = strings.TrimSuffix(args[0], filepath.Ext(args[0])) + "_images"
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outDir}, "would extract embedded images"))
			}
			if err := pdfx.ToImage(context.Background(), args[0], outDir, nil); err != nil {
				return pdfErr("PDF_TO_IMAGE", "check the input PDF", err)
			}
			files, _ := pdfx.ListDirFiles(outDir)
			return done(cmd, newFilesOut(files, "extracted embedded images only; not page rasterization"))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

func pdfFromImage() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "from-image FILE [FILE...]",
		Aliases: []string{"to-pdf"},
		Short:   "Build a PDF from image files",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = "images.pdf"
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, ""))
			}
			if err := pdfx.FromImage(context.Background(), args, outPath); err != nil {
				return pdfErr("PDF_FROM_IMAGE", "check the input images", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

// pdfErr keeps only the root cause of a pdfcpu error chain
// ("optimize: prepare PDF context: read context: xref table: X" becomes "X")
// so the message says what is wrong, not where pdfcpu was.
func pdfErr(code, hint string, err error) error {
	msg := err.Error()
	if i := strings.LastIndex(msg, ": "); i >= 0 {
		msg = msg[i+2:]
	}
	return out.Errf(out.ExitError, code, hint, "%s", msg)
}

func pdfInfo() *cobra.Command {
	return &cobra.Command{
		Use:   "info FILE [FILE...]",
		Short: "Show pages, page size, metadata and encryption of PDFs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			res := infoOut{Files: []pdfx.Info{}}
			for _, f := range args {
				info, err := pdfx.ReadInfo(context.Background(), f)
				if err != nil {
					return pdfErr("PDF_INFO", "check the input PDF", err)
				}
				res.Files = append(res.Files, info)
			}
			return done(cmd, res)
		},
	}
}

type infoOut struct {
	Files []pdfx.Info `json:"files"`
}

func (r infoOut) Text() string {
	var b strings.Builder
	for _, i := range r.Files {
		fmt.Fprintf(&b, "%s  PDF %s, %d pages, %s, %d bytes\n", out.SafeText(i.File), i.Version, i.Pages, i.PageSize, i.SizeBytes)
		for _, kv := range [][2]string{{"title", i.Title}, {"author", i.Author}, {"creator", i.Creator}, {"producer", i.Producer}, {"created", i.Created}, {"modified", i.Modified}} {
			if kv[1] != "" {
				fmt.Fprintf(&b, "  %-9s %s\n", kv[0], out.SafeText(kv[1]))
			}
		}
		fmt.Fprintf(&b, "  encrypted %t, form %t, signatures %t\n", i.Encrypted, i.Form, i.Signatures)
	}
	return b.String()
}

func pdfProtect() *cobra.Command {
	return pdfPasswordCmd("protect", "Encrypt a PDF with a password (AES-256)", ".protected", "PDF_PROTECT", pdfx.Protect)
}

func pdfUnlock() *cobra.Command {
	return pdfPasswordCmd("unlock", "Remove the password from a PDF", ".unlocked", "PDF_UNLOCK", pdfx.Unlock)
}

// pdfPasswordCmd builds protect and unlock. The password comes from a file
// or stdin, never a flag value, so it does not show in the process list or
// shell history.
func pdfPasswordCmd(verb, short, suffix, code string, fn func(context.Context, string, string, string) error) *cobra.Command {
	var pwFile string
	cmd := &cobra.Command{
		Use:   verb + " FILE --password-file FILE",
		Short: short,
		Long: short + `.

The password is read from --password-file (the first line), or from stdin
with --password-file -. It is never taken as a flag value, so it does not
appear in the process list or shell history.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			if pwFile == "" {
				return out.Usage("--password-file is required (a file, or - for stdin)")
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = pdfx.DefaultOut(args[0], suffix, ".pdf")
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, "would write "+outPath))
			}
			pw, err := readPassword(cmd, pwFile)
			if err != nil {
				return err
			}
			if err := fn(context.Background(), args[0], outPath, pw); err != nil {
				return pdfErr(code, "check the password and the input PDF", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	cmd.Flags().StringVar(&pwFile, "password-file", "", "file holding the password, or - for stdin")
	addWriteFlags(cmd)
	return cmd
}

func readPassword(cmd *cobra.Command, path string) (string, error) {
	var r io.Reader = cmd.InOrStdin()
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return "", out.Errf(out.ExitNotFound, "NOT_FOUND", "check --password-file", "%s: no such file", path)
		}
		defer f.Close()
		r = f
	}
	line, err := bufio.NewReader(io.LimitReader(r, 4096)).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", out.Errf(out.ExitError, "PASSWORD", "check --password-file", "read password: %v", err)
	}
	pw := strings.TrimRight(line, "\r\n")
	if pw == "" {
		return "", out.Usage("the password is empty")
	}
	return pw, nil
}

func pdfPages() *cobra.Command {
	var keep, remove string
	cmd := &cobra.Command{
		Use:   "pages FILE (--keep PAGES | --remove PAGES)",
		Short: "Extract, reorder or delete pages of a PDF",
		Long: `Extract, reorder or delete pages of a PDF.

--keep writes only the listed pages, in the order given ("3,1-2" puts page 3
first). --remove writes every page except the listed ones. Pages use the
pdfcpu syntax: "1-3,5", "even", "odd", "l" for the last page.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			if (keep == "") == (remove == "") {
				return out.Usage("give exactly one of --keep or --remove")
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = pdfx.DefaultOut(args[0], ".pages", ".pdf")
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, "would write "+outPath))
			}
			var err error
			if keep != "" {
				err = pdfx.KeepPages(context.Background(), args[0], outPath, keep)
			} else {
				err = pdfx.RemovePages(context.Background(), args[0], outPath, remove)
			}
			if err != nil {
				return pdfErr("PDF_PAGES", "check the page list and the input PDF", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	cmd.Flags().StringVar(&keep, "keep", "", "pages to keep, in this order")
	cmd.Flags().StringVar(&remove, "remove", "", "pages to delete")
	addWriteFlags(cmd)
	return cmd
}

func pdfNumber() *cobra.Command {
	var text, pos, pages string
	cmd := &cobra.Command{
		Use:   "number FILE",
		Short: "Add page numbers to a PDF",
		Long: `Add page numbers to a PDF.

In --text, %p is the page number and %P the page count: "%p of %P" prints
"3 of 10". --position is bl, bc, br (bottom) or tl, tc, tr (top).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFiles(args); err != nil {
				return err
			}
			if !slices.Contains(pdfx.NumberPositions, pos) {
				return out.Usage("--position must be one of %s", strings.Join(pdfx.NumberPositions, ", "))
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = pdfx.DefaultOut(args[0], ".numbered", ".pdf")
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, "would write "+outPath))
			}
			var sel []string
			if pages != "" {
				sel = []string{pages}
			}
			if err := pdfx.NumberPages(context.Background(), args[0], outPath, text, pos, sel); err != nil {
				return pdfErr("PDF_NUMBER", "check --text, --pages and the input PDF", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	cmd.Flags().StringVar(&text, "text", "%p", "number text; %p page, %P page count")
	cmd.Flags().StringVar(&pos, "position", "bc", "bl, bc, br, tl, tc or tr")
	cmd.Flags().StringVar(&pages, "pages", "", "pages to number (empty = all)")
	addWriteFlags(cmd)
	return cmd
}
