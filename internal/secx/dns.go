package secx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/sgnd/toolips/internal/out"
)

// Resolver is the part of net.Resolver that CheckDNS uses; tests pass a fake.
type Resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
}

// DNSCheck is what a domain publishes for web and mail, and what looks wrong.
type DNSCheck struct {
	Domain   string   `json:"domain"`
	A        []string `json:"a"`
	AAAA     []string `json:"aaaa"`
	MX       []MXHost `json:"mx"`
	SPF      []string `json:"spf"`
	DMARC    string   `json:"dmarc,omitempty"`
	Problems []string `json:"problems"`
	NotFound bool     `json:"not_found,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// MXHost is one mail server and its preference (lower is tried first).
type MXHost struct {
	Host string `json:"host"`
	Pref uint16 `json:"pref"`
}

func (d DNSCheck) Text() string {
	if d.Error != "" {
		return fmt.Sprintf("%s  FAIL  %s\n", out.SafeText(d.Domain), out.SafeText(d.Error))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", out.SafeText(d.Domain))
	line := func(name string, vals []string) {
		if len(vals) == 0 {
			fmt.Fprintf(&b, "  %-6s -\n", name)
			return
		}
		for _, v := range vals {
			fmt.Fprintf(&b, "  %-6s %s\n", name, out.SafeText(v))
		}
	}
	line("A", d.A)
	line("AAAA", d.AAAA)
	var mx []string
	for _, m := range d.MX {
		if m.Host == "" {
			mx = append(mx, fmt.Sprintf("%d . (null MX: accepts no mail, on purpose)", m.Pref))
			continue
		}
		mx = append(mx, fmt.Sprintf("%d %s", m.Pref, m.Host))
	}
	line("MX", mx)
	line("SPF", d.SPF)
	var dmarc []string
	if d.DMARC != "" {
		dmarc = []string{d.DMARC}
	}
	line("DMARC", dmarc)
	if len(d.Problems) == 0 {
		b.WriteString("  no problems found\n")
	}
	for _, p := range d.Problems {
		fmt.Fprintf(&b, "  ! %s\n", out.SafeText(p))
	}
	return b.String()
}

// CheckDNS looks up domain with r (the system resolver when nil).
func CheckDNS(domain string, r Resolver, timeout time.Duration) DNSCheck {
	domain = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
	res := DNSCheck{Domain: domain, A: []string{}, AAAA: []string{}, MX: []MXHost{}, SPF: []string{}, Problems: []string{}}
	if domain == "" {
		res.Error = "domain is required"
		return res
	}
	if r == nil {
		r = net.DefaultResolver
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ips, ipErr := r.LookupIPAddr(ctx, domain)
	for _, ip := range ips {
		if ip.IP.To4() != nil {
			res.A = append(res.A, ip.IP.String())
		} else {
			res.AAAA = append(res.AAAA, ip.IP.String())
		}
	}
	mxs, mxErr := r.LookupMX(ctx, domain)
	for _, m := range mxs {
		res.MX = append(res.MX, MXHost{Host: strings.TrimSuffix(m.Host, "."), Pref: m.Pref})
	}
	sort.Slice(res.MX, func(i, j int) bool { return res.MX[i].Pref < res.MX[j].Pref })
	txt, txtErr := r.LookupTXT(ctx, domain)
	for _, t := range txt {
		if strings.HasPrefix(strings.ToLower(t), "v=spf1") {
			res.SPF = append(res.SPF, t)
		}
	}
	dtxt, _ := r.LookupTXT(ctx, "_dmarc."+domain)
	for _, t := range dtxt {
		if strings.HasPrefix(strings.ToLower(t), "v=dmarc1") {
			res.DMARC = t
			break
		}
	}

	// A lookup failure that is not "no such record" means the network or
	// resolver failed; report that instead of a list of missing records.
	for _, err := range []error{ipErr, mxErr, txtErr} {
		if err != nil && !isNotFound(err) {
			res.Error = err.Error()
			return res
		}
	}
	if len(ips) == 0 && len(mxs) == 0 && len(txt) == 0 && isNotFound(ipErr) {
		res.NotFound = true
		res.Error = "no DNS records found for " + domain
		return res
	}
	res.Problems = dnsProblems(res)
	return res
}

func dnsProblems(d DNSCheck) []string {
	p := []string{}
	if len(d.A) == 0 && len(d.AAAA) == 0 {
		p = append(p, "no A or AAAA record: the domain has no web address")
	}
	if len(d.MX) == 0 {
		p = append(p, "no MX record: the domain cannot receive mail")
	}
	switch {
	case len(d.SPF) == 0:
		p = append(p, "no SPF record: anyone can send mail as this domain")
	case len(d.SPF) > 1:
		p = append(p, fmt.Sprintf("%d SPF records: receivers treat more than one as an error", len(d.SPF)))
	}
	switch {
	case d.DMARC == "":
		p = append(p, "no DMARC record at _dmarc."+d.Domain)
	case strings.Contains(strings.ReplaceAll(strings.ToLower(d.DMARC), " ", ""), "p=none"):
		p = append(p, "DMARC policy is p=none: forged mail is reported but still delivered")
	}
	return p
}

func isNotFound(err error) bool {
	var de *net.DNSError
	return errors.As(err, &de) && de.IsNotFound
}

// ResolverFor returns a resolver that asks server ("host:port", port 53 when
// left out) instead of the system's DNS settings.
func ResolverFor(server string) Resolver {
	if _, _, err := net.SplitHostPort(server); err != nil {
		server = net.JoinHostPort(server, "53")
	}
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, server)
		},
	}
}
