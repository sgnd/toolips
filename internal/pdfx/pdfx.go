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

// conf is the pdfcpu configuration for every command. Offline stops pdfcpu
// from fetching links, images or revocation lists, so files stay local.
func conf() *model.Configuration {
	c, err := api.LoadConfiguration(api.ConfigurationOptions{Mode: api.ConfigurationModeStateless})
	if err != nil {
		// Stateless mode should not fail; fall back to a built-in config.
		c = model.NewDefaultConfiguration()
	}
	c.Offline = true
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

// KeepPages writes the selected pages of inFile, in the order given, to
// outFile ("3,1-2" puts page 3 first).
func KeepPages(ctx context.Context, inFile, outFile, sel string) error {
	pages, err := pageList(sel)
	if err != nil {
		return err
	}
	return api.CollectFile(ctx, inFile, outFile, pages, conf())
}

// RemovePages writes inFile without the selected pages to outFile.
func RemovePages(ctx context.Context, inFile, outFile, sel string) error {
	pages, err := pageList(sel)
	if err != nil {
		return err
	}
	return api.RemovePagesFile(ctx, inFile, outFile, pages, conf())
}

func pageList(sel string) ([]string, error) {
	if strings.TrimSpace(sel) == "" {
		return nil, fmt.Errorf("page selection is empty")
	}
	return api.ParsePageSelection(sel)
}

// NumberPositions are the places a page number can go: bottom or top, then
// left, center or right.
var NumberPositions = []string{"bl", "bc", "br", "tl", "tc", "tr"}

// NumberPages stamps a page number on the selected pages. In text, %p is
// the page number and %P the page count ("%p of %P").
func NumberPages(ctx context.Context, inFile, outFile, text, pos string, pages []string) error {
	if text == "" {
		text = "%p"
	}
	ok := false
	for _, p := range NumberPositions {
		ok = ok || p == pos
	}
	if !ok {
		return fmt.Errorf("position must be one of %s (got %q)", strings.Join(NumberPositions, ", "), pos)
	}
	offY := 20
	if pos[0] == 't' {
		offY = -20
	}
	desc := fmt.Sprintf("font:Helvetica, points:10, scale:1 abs, pos:%s, off:0 %d, rot:0, op:1", pos, offY)
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
