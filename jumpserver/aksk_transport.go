package jumpserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// This evaluation fork deliberately has no switch to enable writes.
type signatureTransport struct {
	base   http.RoundTripper
	config *Config
}

func (t *signatureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c := t.config
	if req.Method != http.MethodGet {
		return nil, fmt.Errorf("read-only provider blocks HTTP %s", req.Method)
	}
	if req.URL.Scheme+"://"+req.URL.Host != c.BaseURL || req.URL.User != nil || !strings.HasPrefix(req.URL.Path, "/api/v1/") {
		return nil, fmt.Errorf("read-only provider blocks requests outside the configured API origin and v1 scope")
	}
	r := req.Clone(req.Context())
	r.Header = req.Header.Clone()
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	r.Header.Set("X-JMS-ORG", c.OrganizationID)
	payload := "(request-target): get " + r.URL.RequestURI() + "\naccept: application/json\ndate: " + r.Header.Get("Date")
	mac := hmac.New(sha256.New, []byte(c.AccessKeySecret))
	mac.Write([]byte(payload))
	r.Header.Set("Authorization", fmt.Sprintf(`Signature keyId="%s",algorithm="hmac-sha256",headers="(request-target) accept date",signature="%s"`, c.AccessKeyID, base64.StdEncoding.EncodeToString(mac.Sum(nil))))
	return t.base.RoundTrip(r)
}
