# Releasing

Run from this repository root. Build and test using mise-pinned tools. Preserve the upstream MIT license. Keep internal URLs, real user inventories, credentials, private signing keys and Terraform plans out of this repository and all release packages.

1. Run `go test -race ./...` and `go build -trimpath .`.
2. Create a new semantic version tag. Never overwrite an existing release.
3. Build platform ZIP archives with GoReleaser or the equivalent manual process. Each contains `terraform-provider-jumpserver_vVERSION` and LICENSE. Supported release targets are darwin/linux on amd64/arm64.
4. Publish `terraform-provider-jumpserver_VERSION_manifest.json` with protocol 5.0, a SHA256SUMS covering all ZIPs and the manifest, and a binary detached OpenPGP signature of SHA256SUMS. Sign locally using a dedicated protected signing key. Never upload its private key as a release asset.
5. Upload the release archives, manifest, SHA256SUMS, detached signature and public signing key to the GitHub release. Verify all assets before publishing.
6. Register the repository under the mytxyang namespace in Terraform Registry and add the signing public key. GitHub release publication alone does not publish into Terraform Registry.
7. Verify installation from a clean Terraform directory without dev_overrides. Run only validate/plan against authorized non-production targets.

The supplied GoReleaser configuration creates draft releases. Its signing key and GitHub authorization must be supplied by the operator. No JumpServer AK/SK is needed for release builds.
