---
page_title: "JumpServer Provider"
description: |-
  Read-only JumpServer API v1 queries using AK/SK HTTP Signature authentication.
---
# JumpServer Provider

This independent read-only fork supports existing-user, group and permission inspection and import-plan evaluation. It rejects writes. See the repository README for supported scope and limitations.

## Schema

- `base_url` (String, Required): HTTPS origin without path, query, fragment or embedded credentials.
- `organization_id` (String, Required): Explicit organization UUID sent as X-JMS-ORG.
- `access_key_id` (String, Optional, Sensitive): Defaults to JUMPSERVER_ACCESS_KEY_ID.
- `access_key_secret` (String, Optional, Sensitive): Defaults to JUMPSERVER_ACCESS_KEY_SECRET.
- `skip_tls_verify` (Boolean, Optional): Must remain false.
- `api_version` (String, Deprecated): Retained for schema compatibility; only API v1 is supported.

At least an AK/SK pair must be provided. Password authentication is not supported. Real credentials belong in a secret manager or protected process environment.
