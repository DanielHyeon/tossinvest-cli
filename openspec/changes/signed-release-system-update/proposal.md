## Why

The settings updater can install a pre-staged sibling candidate, but an operator
still needs a development checkout and terminal command to create that candidate.
The console needs a release channel that can obtain the official platform asset
without trusting transport checksums alone or executing remote installer code.

## What Changes

- Add a settings action that discovers the latest stable GitHub release and
  stages the fixed platform asset as the existing sibling `.candidate`.
- Require GitHub Actions/Sigstore build provenance for the exact downloaded
  archive, with the expected repository, release workflow, tag ref, SLSA
  predicate, SHA-256, and transparency-log evidence.
- Bound network redirects, response sizes, archive extraction, and candidate
  publication; preserve the previous candidate on every failure.
- Reuse the existing candidate metadata inspection and human-reviewed install
  action. Download never installs or restarts by itself.
- Publish provenance attestations for release archives from the release workflow.
- Document that only releases created after attestation is enabled are eligible
  for console download.

## Capabilities

### New Capabilities

- `signed-release-distribution`: Defines official release discovery, artifact
  provenance verification, safe extraction, and fixed candidate publication.

### Modified Capabilities

- `operator-console`: Adds an authenticated, CSRF-gated “check and stage signed
  release” settings action while preserving human approval for installation.

## Impact

Affected areas include `.github/workflows/release.yml`, a new signed-release
client/verifier package, `internal/localupdate`, console settings models/routes,
`cmd/tossctl` console wiring, release/update documentation, and Go module
dependencies for Sigstore verification. No account, order, Guardian, journal, or
engine loop behavior changes. Release download requires outbound HTTPS to GitHub
release/attestation endpoints and the Sigstore public-good TUF service.

PM bootstrap exception: this operational hotfix predates an active signed
release-distribution story, so `signed-release-system-update` is temporarily
listed in `docs/pm/portfolio/_registry.yaml` and must be removed when the change
is archived into the normal portfolio hierarchy.
