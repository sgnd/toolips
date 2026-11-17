package cli

import (
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/sgnd/toolips/internal/out"
	"github.com/sgnd/toolips/internal/secx"
)

func cmdTLS() *cobra.Command {
	tls := &cobra.Command{Use: "tls", Short: "TLS checks"}
	var port string
	var timeout time.Duration
	check := &cobra.Command{
		Use:   "check HOST",
		Short: "Check a host's TLS certificate (expiry, chain, protocol)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res := secx.CheckTLS(args[0], port, timeout)
			if !res.OK {
				e := out.Errf(out.ExitNetwork, "TLS_CHECK", "check the host, port and network", "%s", res.Error)
				out.P.Print(commandID(cmd), res, e)
				return exitCode(e.Exit)
			}
			return done(cmd, res)
		},
	}
	check.Flags().StringVar(&port, "port", "443", "TCP port")
	check.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "dial timeout")
	tls.AddCommand(check)
	return tls
}

func cmdHeaders() *cobra.Command {
	headers := &cobra.Command{Use: "headers", Short: "HTTP security headers"}
	check := &cobra.Command{
		Use:   "check URL",
		Short: "Grade security headers on a URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res := secx.CheckHeaders(args[0], nil)
			if res.Error != "" {
				e := out.Errf(out.ExitNetwork, "HEADERS_CHECK", "check the URL and network", "%s", res.Error)
				out.P.Print(commandID(cmd), res, e)
				return exitCode(e.Exit)
			}
			return done(cmd, res)
		},
	}
	headers.AddCommand(check)
	return headers
}

func cmdSecret() *cobra.Command {
	secret := &cobra.Command{Use: "secret", Short: "Secret scanning"}
	scan := &cobra.Command{
		Use:   "scan PATH [PATH...]",
		Short: "Scan files and dirs for keys and tokens (local regex)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := secx.ScanPaths(args)
			if err != nil {
				return out.Errf(out.ExitError, "SECRET_SCAN", "check the paths", "%v", err)
			}
			return done(cmd, res)
		},
	}
	secret.AddCommand(scan)
	return secret
}

func cmdCT() *cobra.Command {
	ct := &cobra.Command{Use: "ct", Short: "Certificate transparency"}
	var baseURL string
	search := &cobra.Command{
		Use:   "search DOMAIN",
		Short: "Search public CT logs for a domain (crt.sh)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := &http.Client{Timeout: 30 * time.Second}
			res := secx.SearchCT(args[0], client, baseURL)
			if res.Error != "" {
				e := out.Errf(out.ExitNetwork, "CT_SEARCH", "check the domain and network", "%s", res.Error)
				out.P.Print(commandID(cmd), res, e)
				return exitCode(e.Exit)
			}
			return done(cmd, res)
		},
	}
	search.Flags().StringVar(&baseURL, "api", "", "CT API base URL (default https://crt.sh/)")
	ct.AddCommand(search)
	return ct
}
