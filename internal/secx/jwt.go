package secx

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sgnd/toolips/internal/out"
)

// JWT is a decoded token. The signature is not checked: toolips has no key,
// and the token never leaves the machine.
type JWT struct {
	Header   map[string]any    `json:"header"`
	Claims   map[string]any    `json:"claims"`
	Times    map[string]string `json:"times,omitempty"` // exp, iat, nbf in RFC 3339
	Expired  bool              `json:"expired"`
	Signed   bool              `json:"signed"`
	Verified bool              `json:"verified"` // always false: see the type comment
}

// DecodeJWT splits and decodes a compact JWT ("header.payload.signature").
func DecodeJWT(token string, now time.Time) (JWT, error) {
	token = strings.TrimSpace(token)
	token = strings.TrimPrefix(token, "Bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JWT{}, errors.New("a JWT has three parts separated by dots")
	}
	var t JWT
	if err := decodePart(parts[0], &t.Header); err != nil {
		return JWT{}, fmt.Errorf("header: %w", err)
	}
	if err := decodePart(parts[1], &t.Claims); err != nil {
		return JWT{}, fmt.Errorf("claims: %w", err)
	}
	t.Signed = parts[2] != ""
	t.Times = map[string]string{}
	for _, k := range []string{"exp", "iat", "nbf"} {
		if v, ok := t.Claims[k].(float64); ok {
			ts := time.Unix(int64(v), 0).UTC()
			t.Times[k] = ts.Format(time.RFC3339)
			if k == "exp" && now.After(ts) {
				t.Expired = true
			}
		}
	}
	return t, nil
}

func decodePart(s string, v any) error {
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
	if err != nil {
		return errors.New("not base64url")
	}
	if err := json.Unmarshal(b, v); err != nil {
		return errors.New("not a JSON object")
	}
	return nil
}

func (t JWT) Text() string {
	var b strings.Builder
	section := func(name string, m map[string]any) {
		fmt.Fprintf(&b, "%s\n", name)
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v, _ := json.Marshal(m[k])
			line := string(v)
			if ts, ok := t.Times[k]; ok && name == "claims" {
				line += "  (" + ts + ")"
			}
			fmt.Fprintf(&b, "  %-10s %s\n", out.SafeText(k), out.SafeText(line))
		}
	}
	section("header", t.Header)
	section("claims", t.Claims)
	if t.Expired {
		b.WriteString("expired\n")
	}
	b.WriteString("signature not checked\n")
	return b.String()
}
