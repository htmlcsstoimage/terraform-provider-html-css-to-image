---
page_title: "Releasing the provider"
---

# Releasing the provider

Releases are driven by the root `VERSION` file. The initial version is `0.1.0-beta.1`. Beta releases are public prereleases; they are not private drafts. Terraform users must select a prerelease explicitly, for example `version = "= 0.1.0-beta.1"`. Use `0.1.0-beta.2` for the next beta and `0.1.0` when ready for the stable release.

## One-time setup

1. Make the `htmlcsstoimage/terraform-provider-html-css-to-image` GitHub repository public and allow its GitHub Actions workflows.
2. Save the ASCII-armored private RSA signing key as the Actions secret `GPG_PRIVATE_KEY`, and its passphrase as `PASSPHRASE`.
3. Register the matching public key in Terraform Registry under the `htmlcsstoimage` namespace.
4. After the first signed GitHub release succeeds, choose **Publish → Provider** in Terraform Registry and select the repository. Subsequent releases are ingested through the Registry's GitHub webhook.

The release workflow uses GitHub's built-in token with `contents: write`; no personal OAuth token is needed.

## Publish a version

Update `VERSION` and push the reviewed changes to `main`. Supported versions are `MAJOR.MINOR.PATCH` and prereleases with `-alpha.N`, `-beta.N`, or `-rc.N`.

After the **Test** workflow succeeds, **Release** checks out that exact tested commit, imports the signing key, creates its `v`-prefixed tag, and runs GoReleaser. It builds ZIP archives for Linux, macOS, Windows, and FreeBSD on amd64 and arm64, embeds the provider version, and uploads the protocol-6 manifest, SHA-256 checksums, and detached GPG signature. The release remains a draft until all uploads succeed. Prerelease tags are marked as prereleases when published.

An already-published version is skipped and never overwritten. If a release fails, rerun its original workflow. An existing unpublished tag must still point to the tested commit; a different commit requires a new version. If partial draft assets prevent a retry, remove only the unpublished draft and rerun the original workflow, preserving its tag. Never replace published artifacts or move published tags.

## Local checks

With GoReleaser v2.18.2 installed:

```sh
goreleaser check
GOWORK=off goreleaser release --snapshot --clean --skip=sign,publish
```

These commands validate packaging and build unsigned local snapshots under `dist/`. They do not test the real signing key or publish anything. CI also validates the GoReleaser configuration. Check the first signed release in Terraform Registry and run `terraform init` without development overrides to verify Registry installation.

See [HashiCorp's publishing requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing).
