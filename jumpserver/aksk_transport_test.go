package jumpserver

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func testConfig() *Config {
	return &Config{BaseURL: "https://example.test", AccessKeyID: "test-id", AccessKeySecret: "test-secret", OrganizationID: "test-org"}
}
func TestSignatureAndIsolation(t *testing.T) {
	c := testConfig()
	calls := 0
	tr := &signatureTransport{config: c, base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("X-JMS-ORG") != "test-org" {
			t.Fatal("org missing")
		}
		msg := "(request-target): get /api/v1/users/users/?username=a%2Bb&limit=2\naccept: application/json\ndate: " + r.Header.Get("Date")
		m := hmac.New(sha256.New, []byte("test-secret"))
		m.Write([]byte(msg))
		want := fmt.Sprintf(`Signature keyId="test-id",algorithm="hmac-sha256",headers="(request-target) accept date",signature="%s"`, base64.StdEncoding.EncodeToString(m.Sum(nil)))
		if r.Header.Get("Authorization") != want {
			t.Fatal("incorrect signature")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
	})}
	r, _ := http.NewRequest("GET", c.BaseURL+"/api/v1/users/users/?username=a%2Bb&limit=2", nil)
	r.Header.Set("Authorization", "Bearer old")
	if _, err := tr.RoundTrip(r); err != nil {
		t.Fatal(err)
	}
	if r.Header.Get("Authorization") != "Bearer old" {
		t.Fatal("mutated caller headers")
	}
	for _, method := range []string{"POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"} {
		r.Method = method
		if _, err := tr.RoundTrip(r); err == nil {
			t.Fatal("write allowed", method)
		}
	}
	for _, target := range []string{"https://evil.test/api/v1/users/", "http://example.test/api/v1/users/", "https://example.test/api/v2/users/", "https://u:p@example.test/api/v1/users/"} {
		r, _ = http.NewRequest("GET", target, nil)
		if _, err := tr.RoundTrip(r); err == nil {
			t.Fatal("origin/scope allowed", target)
		}
	}
	if calls != 1 {
		t.Fatal("blocked request reached network")
	}
	client := c.NewHTTPClient()
	if client.CheckRedirect(r, nil) != http.ErrUseLastResponse {
		t.Fatal("redirect allowed")
	}
}
func TestConfigureWithoutNetwork(t *testing.T) {
	p := Provider()
	if err := p.InternalValidate(); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []map[string]interface{}{
		{"base_url": "https://example.test", "organization_id": "org"},
		{"base_url": "http://example.test", "organization_id": "org", "access_key_id": "a", "access_key_secret": "s"},
		{"base_url": "https://example.test", "organization_id": "org", "access_key_id": "a", "access_key_secret": "s", "skip_tls_verify": true},
	} {
		d := schema.TestResourceDataRaw(t, p.Schema, raw)
		if _, ds := providerConfigure(context.Background(), d); !ds.HasError() {
			t.Fatal("invalid configuration allowed")
		}
	}
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{"base_url": "https://example.test", "organization_id": "org", "access_key_id": "a", "access_key_secret": "s"})
	if _, ds := providerConfigure(context.Background(), d); ds.HasError() {
		t.Fatal(ds)
	}
}
func TestPilotFields(t *testing.T) {
	cases := []struct {
		r         *schema.Resource
		key, attr string
	}{{resourceUser(), "system_roles", "id"}, {resourceUserGroup(), "users", "id"}, {resourceAssetPermission(), "users", "id"}, {resourceAssetPermission(), "assets", "id"}, {resourceAssetPermission(), "actions", "value"}}
	for _, c := range cases {
		d := schema.TestResourceDataRaw(t, c.r.Schema, nil)
		data := map[string]interface{}{c.key: []interface{}{map[string]interface{}{c.attr: "one"}, "two"}}
		if err := setPilotList(d, data, c.key); err != nil {
			t.Fatal(err)
		}
		v := d.Get(c.key).([]interface{})
		if len(v) != 2 || v[0] != "one" || v[1] != "two" {
			t.Fatal(v)
		}
		if err := setPilotList(d, map[string]interface{}{c.key: []interface{}{map[string]interface{}{"label": "bad"}}}, c.key); err == nil {
			t.Fatal("malformed field silently accepted")
		}
	}
}

func TestLookupExactAndEncoded(t *testing.T) {
	payload := `{"count":1,"results":[{"id":"u","username":"a+b"}]}`
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("username") != "a+b" || r.URL.Query().Get("limit") != "2" {
			t.Error("query not encoded")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, payload)
	}))
	defer server.Close()
	c := testConfig()
	c.BaseURL = server.URL
	// Only this local fixture uses the test certificate; configure rejects this in live use.
	c.SkipTLSVerify = true
	if _, err := c.dataSourceLookup("users/users/", "username", "a+b"); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{`{"count":2,"results":[{"username":"a+b"}]}`, `[{"username":"other"}]`, `[]`, `[{"username":"a+b"},{"username":"a+b"}]`} {
		payload = bad
		if _, err := c.dataSourceLookup("users/users/", "username", "a+b"); err == nil {
			t.Fatal("ambiguous or wrong user accepted")
		}
	}
}
