package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/sgnd/toolips/internal/out"
	"github.com/sgnd/toolips/internal/secx"
)

func cmdJWT() *cobra.Command {
	j := &cobra.Command{Use: "jwt", Short: "JSON Web Tokens"}
	decode := &cobra.Command{
		Use:   "decode [FILE]",
		Short: "Show a JWT's header and claims, with readable exp, iat and nbf times",
		Long: `Show a JWT's header and claims, with readable exp, iat and nbf times.

The token is read from FILE, or from stdin when FILE is left out or is "-",
never from an argument, so it does not show in the process list or shell
history. A leading "Bearer " is removed. The signature is not checked, and
the token never leaves this machine.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "-"
			if len(args) == 1 {
				path = args[0]
			}
			tok, err := readSecretLine(cmd, path, "token", "pass a file holding the token, or pipe it on stdin", 64<<10)
			if err != nil {
				return err
			}
			res, err := secx.DecodeJWT(tok, time.Now())
			if err != nil {
				return out.Errf(out.ExitError, "JWT", "pass a compact JWT: header.payload.signature", "%v", err)
			}
			return done(cmd, res)
		},
	}
	j.AddCommand(decode)
	return j
}
