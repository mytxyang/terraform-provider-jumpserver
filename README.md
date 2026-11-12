# JumpServer Terraform Provider — AK/SK read-only edition

A deliberately read-only provider for inspecting JumpServer users, user groups and asset permissions with HTTP Signature authentication.

Derived from [astronas/terraform-provider-jumpserver](https://github.com/astronas/terraform-provider-jumpserver), version 1.6.1, commit `b98918fe1c53d912eaa2d673e399cc3d0493e205`. The upstream MIT license and copyright are retained in LICENSE. This is an independent fork, not an official JumpServer release.

## Scope

Version 0.1.0 only sends HTTPS GET requests to API v1 at the configured origin. Password login, mutation requests, cross-origin requests, automatic redirects and disabled TLS verification are rejected. There is no switch to enable writes.

Supported data sources: `jumpserver_user`, `jumpserver_user_group`, `jumpserver_asset_permission`. Resources with the same names are exposed for reading existing objects and evaluating declarative import plans. Create, update and delete are unsupported. Do not run apply against this evaluation release: even an import-only apply can persist Terraform state.

## Usage

```hcl
terraform {
  required_providers {
    jumpserver = {
      source  = "mytxyang/jumpserver"
      version = "= 0.1.0"
    }
  }
}

provider "jumpserver" {
  base_url        = "https://jumpserver.example.com"
  organization_id = "00000000-0000-0000-0000-000000000002"
}

data "jumpserver_user" "example" {
  username = "example-user"
}
```

Supply `JUMPSERVER_ACCESS_KEY_ID` and `JUMPSERVER_ACCESS_KEY_SECRET` from your secret manager or a protected process environment. Never place real credentials in examples, commits or CI logs. No live credentials are required to build or run offline tests.

After installation: `terraform init`, `terraform validate`, then `terraform plan`. Registry publication is separate from creating a GitHub release; if the Registry has not indexed the release, use an installation mirror with checksum verification.

## Limitations

The access key retains its owner's server-side permissions; client restrictions do not reduce those permissions elsewhere. User system roles are IDs. Missing organization roles are unknown, not evidence of no access. Effective permissions also depend on group membership, organization scope, role bindings and validity dates.

The resource schemas do not model every server property, including all permission date/protocol attributes. No-change plans only cover modeled fields. List ordering can create apparent drift. This release does not claim CRUD, post-import idempotency or complete access-audit coverage.

## Development

Use Go 1.27.4 and Terraform 1.15.6 through mise. Run `go test -race ./...` and `go build -trimpath .`. Offline tests validate signatures, request isolation, mutation blocking, configuration validation, field normalization and exact lookups. Run live checks only with explicit authorization and protected credentials.

See [provider documentation](docs/index.md) and [release instructions](RELEASING.md).
