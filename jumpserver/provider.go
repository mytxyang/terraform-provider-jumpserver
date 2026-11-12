package jumpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	OrganizationID  string
	Token           string
	BaseURL         string
	Username        string
	Password        string
	SkipTLSVerify   bool
	APIVersion      string
}

func (c *Config) NewHTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.SkipTLSVerify,
		},
	}
	return &http.Client{Transport: &signatureTransport{base: transport, config: c}, Timeout: 30 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
}

// GetAPIEndpoint returns the API v1 endpoint URL (used by most JumpServer resources)
func (c *Config) GetAPIEndpoint(path string) string {
	return fmt.Sprintf("%s/api/v1/%s", c.BaseURL, path)
}

// GetAPIEndpointV2 returns the API v2 endpoint URL (used by perms and other v2-only resources)
func (c *Config) GetAPIEndpointV2(path string) string {
	return fmt.Sprintf("%s/api/v2/%s", c.BaseURL, path)
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access_key_id":     {Type: schema.TypeString, Optional: true, Sensitive: true, DefaultFunc: schema.EnvDefaultFunc("JUMPSERVER_ACCESS_KEY_ID", nil)},
			"access_key_secret": {Type: schema.TypeString, Optional: true, Sensitive: true, DefaultFunc: schema.EnvDefaultFunc("JUMPSERVER_ACCESS_KEY_SECRET", nil)},
			"organization_id":   {Type: schema.TypeString, Required: true},

			"skip_tls_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, skip SSL certificate validation (insecure).",
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "v1",
				Deprecated:  "This field is deprecated and no longer affects API routing. Each resource now uses its correct API version automatically.",
				Description: "Deprecated. JumpServer API version (v1 or v2). No longer used — each resource targets its own API version.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"jumpserver_user":             dataSourceUser(),
			"jumpserver_user_group":       dataSourceUserGroup(),
			"jumpserver_asset_permission": dataSourceAssetPermission(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"jumpserver_user":             resourceUser(),
			"jumpserver_user_group":       resourceUserGroup(),
			"jumpserver_asset_permission": resourceAssetPermission(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	base := strings.TrimRight(d.Get("base_url").(string), "/")
	u, err := url.Parse(base)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
		return nil, diag.Errorf("base_url must be an HTTPS origin")
	}
	if d.Get("skip_tls_verify").(bool) {
		return nil, diag.Errorf("TLS verification is mandatory in this evaluation fork")
	}
	ak, sk := d.Get("access_key_id").(string), d.Get("access_key_secret").(string)
	org := d.Get("organization_id").(string)
	if ak == "" || sk == "" || org == "" || strings.ContainsAny(ak, "\"\r\n") || strings.ContainsAny(org, "\r\n") {
		return nil, diag.Errorf("valid AK/SK and organization_id are required; password login is disabled")
	}
	return &Config{BaseURL: base, AccessKeyID: ak, AccessKeySecret: sk, OrganizationID: org, APIVersion: "v1"}, nil
}
