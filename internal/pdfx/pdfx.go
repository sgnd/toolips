// Package pdfx wraps pdfcpu for merge, split, compress, rotate, watermark
// and image import/extract. Pure Go (CGO_ENABLED=0).
package pdfx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func conf() *model.Configuration {
	c, err := api.LoadConfiguration(api.ConfigurationOptions{Mode: api.ConfigurationModeStateless})
	if err != nil {
		// Stateless mode should not fail; fall back to a built-in config.
		return model.NewDefaultConfiguration()
	}
	return c
}

// Merge concatenates inFiles into outFile.
func Merge(ctx context.Context, inFiles []string, outFile string) error {
	if len(inFiles) < 2 {
		return fmt.Errorf("need at least two PDF files to merge")
	}
	if err := ensureParent(outFile); err != nil {
		return err
	}
	return api.MergeCreateFile(ctx, inFiles, outFile, false, conf())
}

// Split writes one PDF per span pages into outDir. span 1 means one page
// per file.
func Split(ctx context.Context, inFile, outDir string, span int) error {
	if span < 1 {
		span = 1
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	return api.SplitFile(ctx, inFile, outDir, span, conf())
}

// Compress optimizes inFile into outFile (pdfcpu Optimize).
func Compress(ctx context.Context, inFile, outFile string) error {
	if err := ensureParent(outFile); err != nil {
		return err
	}
	return api.OptimizeFile(ctx, inFile, outFile, conf(), nil)
}

// Rotate rotates selected pages clockwise by degrees (90/180/270).
// An empty outFile overwrites in place via a temp file.
func Rotate(ctx context.Context, inFile, outFile string, degrees int, pages []string) error {
	if degrees%90 != 0 {
		return fmt.Errorf("rotation must be a multiple of 90 (got %d)", degrees)
	}
	if outFile == "" {
		outFile = inFile
	}
	if err := ensureParent(outFile); err != nil {
		return err
	}
	return api.RotateFile(ctx, inFile, outFile, degrees, pages, conf())
}

// Watermark adds a text watermark to inFile, writing outFile.
func Watermark(ctx context.Context, inFile, outFile, text, desc string, pages []string) error {
	if text == "" {
		return fmt.Errorf("watermark text is required")
	}
	if desc == "" {
		desc = "font:Helvetica, points:48, scale:0.5, op:0.3, rot:45"
	}
	if outFile == "" {
		outFile = inFile
	}
	if err := ensureParent(outFile); err != nil {
		return err
	}
	c := conf()
	wm, err := api.TextWatermark(ctx, text, desc, true, false, types.POINTS, c)
	if err != nil {
		return err
	}
	return api.AddWatermarksFile(ctx, inFile, outFile, pages, wm, c)
}

// ToImage extracts embedded images from inFile into outDir.
// This is not page rasterization (that would need CGO); it only pulls
// images already embedded in the PDF.
func ToImage(ctx context.Context, inFile, outDir string, pages []string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	return api.ExtractImagesFile(ctx, inFile, outDir, pages, conf())
}

// FromImage builds a PDF from image files (pdfcpu ImportImages).
func FromImage(ctx context.Context, imgFiles []string, outFile string) error {
	if len(imgFiles) == 0 {
		return fmt.Errorf("need at least one image")
	}
	if err := ensureParent(outFile); err != nil {
		return err
	}
	return api.ImportImagesFile(ctx, imgFiles, outFile, nil, conf())
}

// ListDirFiles returns non-directory files under dir (non-recursive).
func ListDirFiles(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		out = append(out, filepath.Join(dir, e.Name()))
	}
	return out, nil
}

func ensureParent(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// DefaultOut beside inPath with suffix before the extension.
func DefaultOut(inPath, suffix, ext string) string {
	dir := filepath.Dir(inPath)
	base := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath))
	if ext == "" {
		ext = filepath.Ext(inPath)
	}
	if !strings.HasPrefix(ext, ".") && ext != "" {
		ext = "." + ext
	}
	return filepath.Join(dir, base+suffix+ext)
}
