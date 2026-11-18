package cli

import (
	"context"
	"path/filepath"
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
		pdfWatermark(), pdfToImage(), pdfFromImage(),
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
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, "would merge "+strconv.Itoa(len(args))+" files"))
			}
			if err := pdfx.Merge(context.Background(), args, outPath); err != nil {
				return out.Errf(out.ExitError, "PDF_MERGE", "check the input PDFs", "%v", err)
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
				return out.Errf(out.ExitError, "PDF_SPLIT", "check the input PDF", "%v", err)
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
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, ""))
			}
			if err := pdfx.Compress(context.Background(), args[0], outPath); err != nil {
				return out.Errf(out.ExitError, "PDF_COMPRESS", "check the input PDF", "%v", err)
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
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, ""))
			}
			if err := pdfx.Rotate(context.Background(), args[0], outPath, degrees, sel); err != nil {
				return out.Errf(out.ExitError, "PDF_ROTATE", "use --degrees 90, 180 or 270", "%v", err)
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
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, ""))
			}
			if err := pdfx.Watermark(context.Background(), args[0], outPath, text, desc, sel); err != nil {
				return out.Errf(out.ExitError, "PDF_WATERMARK", "check --text and the input PDF", "%v", err)
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
				return out.Errf(out.ExitError, "PDF_TO_IMAGE", "check the input PDF", "%v", err)
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
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, ""))
			}
			if err := pdfx.FromImage(context.Background(), args, outPath); err != nil {
				return out.Errf(out.ExitError, "PDF_FROM_IMAGE", "check the input images", "%v", err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	addWriteFlags(cmd)
	return cmd
}
