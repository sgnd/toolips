package cli

import (
	"runtime"

	"github.com/spf13/cobra"
)

func cmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print toolips version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			data := versionOut{
				Version: Version,
				Go:      runtime.Version(),
				OS:      runtime.GOOS,
				Arch:    runtime.GOARCH,
			}
			return done(cmd, data)
		},
	}
}

type versionOut struct {
	Version string `json:"version"`
	Go      string `json:"go"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

func (v versionOut) Text() string {
	return "toolips " + v.Version + " (" + v.Go + " " + v.OS + "/" + v.Arch + ")\n"
}
