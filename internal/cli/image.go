package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sgnd/toolips/internal/imgx"
	"github.com/sgnd/toolips/internal/out"
)

func cmdImage() *cobra.Command {
	img := &cobra.Command{Use: "image", Short: "Image tools"}
	img.AddCommand(
		imageCompress(), imageResize(), imageConvert(), imageCrop(),
		imageRotate(), imageWatermark(), imageGrayscale(),
	)
	return img
}

func imageCompress() *cobra.Command {
	var quality int
	cmd := &cobra.Command{
		Use:   "compress FILE [FILE...]",
		Short: "Re-encode images (JPEG quality)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImageBatch(cmd, args, ".compressed", imgx.FormatJPEG, func(in, dest string) error {
				img, _, err := imgx.Open(in)
				if err != nil {
					return err
				}
				return imgx.Save(dest, imgx.Compress(img, quality), imgx.FormatJPEG, quality)
			})
		},
	}
	cmd.Flags().IntVar(&quality, "quality", 75, "JPEG quality 1–100")
	addWriteFlags(cmd)
	return cmd
}

func imageResize() *cobra.Command {
	var width, height int
	cmd := &cobra.Command{
		Use:   "resize FILE [FILE...]",
		Short: "Resize images (keeps aspect if one side is 0)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if width <= 0 && height <= 0 {
				return out.Usage("set --width and/or --height")
			}
			return runImageBatch(cmd, args, ".resized", "", func(in, dest string) error {
				img, format, err := imgx.Open(in)
				if err != nil {
					return err
				}
				if f, err := imgx.ExtFormat(dest); err == nil {
					format = f
				}
				return imgx.Save(dest, imgx.Resize(img, width, height), format, 85)
			})
		},
	}
	cmd.Flags().IntVar(&width, "width", 0, "target width (0 keeps aspect)")
	cmd.Flags().IntVar(&height, "height", 0, "target height (0 keeps aspect)")
	addWriteFlags(cmd)
	return cmd
}

func imageConvert() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "convert FILE [FILE...]",
		Short: "Convert images to jpg, png or gif",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmtOut, err := imgx.ParseFormat(format)
			if err != nil {
				return out.Usage("%v", err)
			}
			return runImageBatch(cmd, args, ".converted", fmtOut, func(in, dest string) error {
				img, _, err := imgx.Open(in)
				if err != nil {
					return err
				}
				return imgx.Save(dest, img, fmtOut, 85)
			})
		},
	}
	cmd.Flags().StringVar(&format, "format", "png", "output format: jpg, png, gif")
	addWriteFlags(cmd)
	return cmd
}

func imageCrop() *cobra.Command {
	var x, y, w, h int
	cmd := &cobra.Command{
		Use:   "crop FILE [FILE...]",
		Short: "Crop images to a rectangle",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if w <= 0 || h <= 0 {
				return out.Usage("--width and --height must be positive")
			}
			return runImageBatch(cmd, args, ".cropped", "", func(in, dest string) error {
				img, format, err := imgx.Open(in)
				if err != nil {
					return err
				}
				cropped, err := imgx.Crop(img, x, y, w, h)
				if err != nil {
					return err
				}
				if f, err := imgx.ExtFormat(dest); err == nil {
					format = f
				}
				return imgx.Save(dest, cropped, format, 85)
			})
		},
	}
	cmd.Flags().IntVar(&x, "x", 0, "left")
	cmd.Flags().IntVar(&y, "y", 0, "top")
	cmd.Flags().IntVar(&w, "width", 0, "crop width")
	cmd.Flags().IntVar(&h, "height", 0, "crop height")
	addWriteFlags(cmd)
	return cmd
}

func imageRotate() *cobra.Command {
	var degrees int
	cmd := &cobra.Command{
		Use:   "rotate FILE [FILE...]",
		Short: "Rotate images 90/180/270 clockwise",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImageBatch(cmd, args, ".rotated", "", func(in, dest string) error {
				img, format, err := imgx.Open(in)
				if err != nil {
					return err
				}
				rot, err := imgx.Rotate(img, degrees)
				if err != nil {
					return err
				}
				if f, err := imgx.ExtFormat(dest); err == nil {
					format = f
				}
				return imgx.Save(dest, rot, format, 85)
			})
		},
	}
	cmd.Flags().IntVar(&degrees, "degrees", 90, "clockwise rotation")
	addWriteFlags(cmd)
	return cmd
}

func imageWatermark() *cobra.Command {
	var text string
	cmd := &cobra.Command{
		Use:   "watermark FILE [FILE...]",
		Short: "Draw text watermark on images",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if text == "" {
				return out.Usage("--text is required")
			}
			return runImageBatch(cmd, args, ".wm", "", func(in, dest string) error {
				img, format, err := imgx.Open(in)
				if err != nil {
					return err
				}
				if f, err := imgx.ExtFormat(dest); err == nil {
					format = f
				}
				return imgx.Save(dest, imgx.Watermark(img, text), format, 85)
			})
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "watermark text")
	addWriteFlags(cmd)
	return cmd
}

func imageGrayscale() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grayscale FILE [FILE...]",
		Short: "Convert images to grayscale",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImageBatch(cmd, args, ".gray", "", func(in, dest string) error {
				img, format, err := imgx.Open(in)
				if err != nil {
					return err
				}
				if f, err := imgx.ExtFormat(dest); err == nil {
					format = f
				}
				return imgx.Save(dest, imgx.Grayscale(img), format, 85)
			})
		},
	}
	addWriteFlags(cmd)
	return cmd
}

func runImageBatch(cmd *cobra.Command, args []string, suffix string, forceFmt imgx.Format, fn func(in, dest string) error) error {
	expanded, err := expandImageInputs(args)
	if err != nil {
		return out.Errf(out.ExitNotFound, "IMAGE", "pass image files, a folder, or a glob (*.jpg|*.png|*.gif)", "%v", err)
	}
	outPath, outDir := writeFlag(cmd)
	if len(expanded) > 1 && outPath != "" {
		return out.Usage("--out is for a single file; use --outdir for batch")
	}
	var written []string
	for _, in := range expanded {
		dest, err := resolveImageOut(in, outPath, outDir, suffix, forceFmt)
		if err != nil {
			return out.Errf(out.ExitError, "IMAGE", "check the input image", "%s: %v", in, err)
		}
		if g.dryRun {
			written = append(written, dest)
			continue
		}
		if err := fn(in, dest); err != nil {
			return out.Errf(out.ExitError, "IMAGE", "check the input image and flags", "%s: %v", in, err)
		}
		written = append(written, dest)
	}
	note := ""
	if g.dryRun {
		note = "would write " + strconv.Itoa(len(written)) + " files"
	}
	return done(cmd, newFilesOut(written, note))
}

func resolveImageOut(in, outPath, outDir, suffix string, forceFmt imgx.Format) (string, error) {
	if outPath != "" {
		return outPath, nil
	}
	format := forceFmt
	if format == "" {
		var err error
		format, err = imgx.ExtFormat(in)
		if err != nil {
			_, format, err = imgx.Open(in)
			if err != nil {
				return "", err
			}
		}
	}
	name := filepath.Base(imgx.DefaultOut(filepath.Base(in), suffix, format))
	if outDir != "" {
		return filepath.Join(outDir, name), nil
	}
	return filepath.Join(filepath.Dir(in), name), nil
}

// expandImageInputs turns globs and directories into concrete JPG/PNG/GIF paths.
// Agents often pass literal "*.jpg"; shells may already expand, so plain files
// are kept as-is.
func expandImageInputs(args []string) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string
	add := func(paths ...string) {
		for _, p := range paths {
			if p == "" {
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	for _, arg := range args {
		if strings.ContainsAny(arg, "*?[") {
			matches, err := filepath.Glob(arg)
			if err != nil {
				return nil, err
			}
			var images []string
			for _, m := range matches {
				if isImagePath(m) {
					images = append(images, m)
				}
			}
			if len(images) == 0 {
				return nil, fmt.Errorf("no images matched %q", arg)
			}
			sort.Strings(images)
			add(images...)
			continue
		}
		st, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		if st.IsDir() {
			entries, err := os.ReadDir(arg)
			if err != nil {
				return nil, err
			}
			var images []string
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				p := filepath.Join(arg, e.Name())
				if isImagePath(p) {
					images = append(images, p)
				}
			}
			if len(images) == 0 {
				return nil, fmt.Errorf("no images in folder %q", arg)
			}
			sort.Strings(images)
			add(images...)
			continue
		}
		add(arg)
	}
	if len(out) == 0 {
		return nil, os.ErrNotExist
	}
	return out, nil
}

func isImagePath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	default:
		return false
	}
}
