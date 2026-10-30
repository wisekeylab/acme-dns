# AGENTS.md — wisekeylab/acme-dns

Instructions for anyone, human or AI coding agent, changing this repository.

## What this fork is

A fork of [acme-dns](https://github.com/acme-dns/acme-dns), maintained by WISeKey. It exists for **one** reason: upstream accepts a TXT value only if it is exactly 43 characters long (the length of an ACME DNS-01 key authorization digest), and some certificate authorities produce domain-validation values of other lengths. This fork relaxes that length check so those values can be published.

## Rules

1. **Keep the difference from upstream minimal.** The only intended functional difference is the TXT value length check. Any other change belongs upstream first; do not add features here.
2. **Offer the fix upstream.** The length change is proposed to upstream acme-dns. If upstream accepts an equivalent change, this fork can be retired.
3. **Stay in step with upstream.** Merge upstream releases into `master` rather than letting the fork drift.
4. **Changes go through pull requests.** Do not push to `master` directly. A maintainer reviews and merges; an AI agent never merges.
5. **Tests must pass:** `go test ./...` and `golangci-lint run`, as the inherited workflows run them. Add a test with every behaviour change.
6. **This repository is public.** Never commit secrets, credentials, internal documents, internal ticket references or links to private repositories.

## Releases

**Do not push release tags yet.** The inherited `release.yml` workflow publishes to upstream's Docker Hub namespace and signs with keys this fork does not hold, so a tag would fail. Where this fork's container image is published, and whether it is signed, is still to be decided; this section will say so once it is.

## Building and testing

```sh
go build ./...
go test ./...
golangci-lint run
docker build -t acme-dns:local .
```
