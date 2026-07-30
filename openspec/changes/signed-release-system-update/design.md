## Context

The existing settings updater is deliberately local: it inspects and installs
only `<running executable>.candidate`, never accepts an HTTP-selected path, and
reuses the engine/verification flock plus same-port relaunch. Development
staging is safe but requires a checkout and `make stage-local-update`.

The legacy `tossctl update` binary-install path executes a remotely fetched
installer and validates only a checksum fetched from the same release channel.
It is not suitable as the backend of a browser action. GitHub Actions artifact
attestations provide keyless Sigstore signatures whose certificate identifies
the source repository, workflow, and ref and whose transparency-log evidence
binds the release archive digest.

## Goals / Non-Goals

**Goals:**

- Let an authenticated operator check and stage the latest stable official
  release entirely from settings.
- Cryptographically bind the downloaded archive to the repository's release
  workflow and exact tag before extracting any executable.
- Publish only a validated `cmd/tossctl` binary to the existing fixed candidate
  path, atomically and without damaging an earlier candidate on failure.
- Keep the existing separate human click, idle-only replacement, rollback, and
  same-port restart.

**Non-Goals:**

- Automatic installation, restart, polling, or background download.
- Remote selection of repository, URL, tag, asset, destination, or command.
- Replacing Homebrew or the legacy CLI update command in this change.
- Updating `auth-helper`; this change stages the running Go executable only.
- Supporting unsigned historical releases.

## Decisions

### D1. GitHub Actions keyless provenance is the signature

Each release build uses `actions/attest@v4` for the completed platform archive
with `id-token: write` and `attestations: write`. The console verifies with the
pinned `sigstore-go` module. Its audited public-good `root.json` is embedded in
that module and TossOS also pins its expected SHA-256, so a dependency update
cannot silently replace the initial trust anchor. TUF metadata is refreshed into
a TossOS-owned mode-0700 cache and the TUF client must successfully enforce
metadata signatures, expiry, consistent snapshots, version monotonicity, and
rollback/freeze protection before a Sigstore trusted root is accepted.

The bundle is accepted only after `sigstore-go`'s complete verification policy
succeeds. That policy verifies the Fulcio chain and SCT, Rekor signed inclusion
proof/checkpoint, the integrated timestamp and certificate validity at that
time, plus:

- the archive's locally computed SHA-256;
- at least one transparency-log entry and integrated timestamp;
- OIDC issuer `https://token.actions.githubusercontent.com`;
- exact SAN
  `https://github.com/JungHoonGhae/tossinvest-cli/.github/workflows/release.yml@refs/tags/<tag>`;
- SLSA provenance v1; and
- a subject digest equal to the downloaded archive.

Presence of JSON fields is never treated as verification. The SLSA predicate
must also bind the GitHub Actions source repository, `release.yml`, exact tag
ref, and resolved source commit. The extracted Go binary's VCS revision must
equal that resolved commit and must not be marked modified.

This avoids a long-lived repository signing key and its provisioning/rotation
problem. A checksum beside an asset is retained for other installers but is not
accepted as the console's signature.

Alternatives rejected:

- checksum-only: the asset and checksum can be replaced together;
- `curl | sh`: executes unauthenticated remote installer logic;
- a repository Ed25519 key: requires creating, storing, rotating, and recovering
  a high-value private secret;
- shelling out to `gh attestation verify`: makes PATH/external binary identity a
  new trust root and inherits CLI-version vulnerabilities.

### D2. The HTTP surface selects nothing

Settings adds `POST /settings/system-update/download`. It accepts only the
normal session and CSRF token. Repository, API base, release tag, platform,
asset name, destination, signer workflow, and trust root are constructor-bound.
Unknown form fields do not influence those values.

The server fetches GitHub's latest stable release record, requires a canonical
stable `vMAJOR.MINOR.PATCH` tag, rejects drafts/prereleases, and selects exactly:

- `tossctl-<goos>-<goarch>.tar.gz` on supported Unix systems;
- no downloadable candidate on platforms where the local replacement path is
  unsupported.

For a canonical current release version, the discovered tag must be strictly
newer; a validly signed older or equal release is refused. A locally built
`dev`/non-release binary is the one bootstrap exception: it may stage a signed
stable release, and the UI explicitly says semantic ordering was unavailable.
The candidate's source commit still has to match the attested tag build.

Tests inject an `httptest` source and verifier; production constructors alone
carry GitHub/Sigstore endpoints.

### D3. Network and archive handling are bounded and fail closed

Metadata, archive, attestation-list, bundle, and TUF requests have context
deadlines and response-size limits. `Content-Length` is an early rejection hint;
bounded readers remain authoritative so legitimate chunked responses work.
TUF uses an injected context-aware fetcher rather than go-tuf's default
synchronous fetcher: it accepts only exact
`https://tuf-repo-cdn.sigstore.dev:443`, follows no redirect, applies both the
caller cancellation and a 30-second client timeout, and caps every response at
16 MiB in addition to TUF's metadata-derived length.
Redirects are limited to five, remain HTTPS, reject userinfo and non-443 ports,
and are restricted to exact GitHub API/release/CDN hosts. Attestation bundle
URLs are accepted only from the exact GitHub hosts or the exact production
attestation host `tmaproduction.blob.core.windows.net` with an
`/attestations/` path. Bundle delivery never follows a redirect; a GitHub
delivery-host change fails closed until reviewed. Private/literal IP hosts and
suffix tricks are rejected. Public requests carry no Authorization header.
Attestation pagination, bundles per digest, pages, and aggregate bytes are
bounded, and every candidate bundle is independently verified. Non-2xx
responses, excess bytes, wrong content, or verification failures never publish
a candidate.

The gzip reader disables multistream input. The tar reader never writes archive
paths. It caps compressed bytes, total uncompressed bytes, header count, and the
binary size. It copies only the exact regular entry `tossctl`; rejects duplicate
entries, sparse/PAX/GNU path tricks, NUL/backslash/absolute/`..` names, links,
devices and FIFOs; and permits only zero-size directory headers and bounded
regular files at `auth-helper` or below `auth-helper/` to be ignored. Extraction
occurs only after the archive attestation succeeds.

### D4. `localupdate` owns fixed candidate publication

`localupdate.Updater` gains a staging method that accepts verified executable
bytes, writes a same-directory no-follow temporary regular file, syncs it,
validates SHA-256/build info/module/command/GOOS/GOARCH without execution, and
atomically publishes it as the updater's already-fixed `.candidate`.

The updater serializes `Inspect`, staging, and installation internally. Staging
copies any existing regular no-follow candidate to a same-directory recovery
temporary, syncs that copy, then atomically renames the fully validated new
temporary over `.candidate` and syncs the directory. If that final sync fails it
restores from a second copy of the recovery file and syncs again. If restoration
itself cannot complete, the operation reports both errors and retains the
recovery file path instead of claiming the old candidate was restored; the
running executable is never involved.

The entire signed-release operation—discovery, TUF refresh, bundle verification,
extraction, and final publish—is serialized by a dedicated release mutex. It is
separate from the console activity exclusion, so a stalled network cannot delay
engine or verification start, and it also prevents concurrent TUF cache writes
or an older slow download from publishing after a newer fast download.
Only the final local validation/rename window takes the existing activity
exclusion. Immediately before publish the handler rechecks that no install
commit is in progress.

If a previous candidate exists, it remains available until the new temporary
file is fully validated. Failures before rename leave it untouched; failures
after rename restore it or retain an explicit recovery artifact as just
described. Tests cover rename, first sync, restore rename, and second sync
failures without overstating power-loss durability.
Install continues to re-open, copy, inspect, and bind to the operator-reviewed
SHA, so staging does not weaken the existing time-of-check/time-of-use defense.

### D5. Download and install remain separate human actions

A successful download redirects to settings with a process-local notice
containing the release tag, asset name, archive SHA-256, signer workflow, and
new candidate SHA-256. The normal candidate table then provides the existing
reviewed install button. Download never calls `Install`, never requests
relaunch, and never stops an engine or verification.

After a console restart the candidate is still safely inspectable as a local
candidate, but the console does not claim remembered provenance unless it
re-verifies it. This change avoids introducing a durable provenance sidecar
whose authenticity would itself need another trust and rollback protocol.

## Risks / Trade-offs

- [Sigstore/TUF adds dependencies and first-use network latency] → pin the Go
  module and embedded-root digest, use a TossOS-owned mode-0700 cache directory,
  bound every request, and surface actionable failure text without touching the
  current candidate.
- [A release predating attestations cannot be staged] → fail closed and state
  that the next attested release is required.
- [GitHub/Sigstore outage] → keep the current binary and previous candidate;
  local `make stage-local-update` remains available.
- [Mutable GitHub release asset] → verify the actual downloaded digest against
  workflow identity and transparency evidence, not the release page or sibling
  checksum.
- [Archive decompression bomb] → cap archive bytes, extracted bytes, entry count,
  and accept only one exact regular binary entry.
- [Download races with install] → keep network work outside the activity lock,
  serialize signed downloads with one another, serialize only candidate
  publication with install/start transitions, and preserve the installer's
  descriptor/hash checks.
- [A valid historical attestation is replayed] → require a strictly newer tag
  for release builds, exact tag/workflow/source-ref provenance, and candidate
  VCS revision equality; allow only an explicit non-release bootstrap.

### D6. The release workflow itself is immutable enough to trust

All third-party actions in the release path are pinned to full commit SHAs.
Repository-level permissions default to none; build jobs receive only
`contents: read`, `id-token: write`, `attestations: write`, and the
`artifact-metadata: write` required by `actions/attest@v4`, while the publishing
job alone receives `contents: write`. A canonical tag guard runs before
building, and each completed archive is attested before release publication.

An existing release/tag is immutable: the workflow does not use `--clobber`,
does not replace an asset, and does not edit an attested release on rerun. A
changed build requires a new patch tag. This is enforced by a static workflow
test. The legacy `tossctl update` command remains for compatibility but prints a
deprecation warning that it is checksum-only and directs operators to the signed
settings flow; replacing that command is a required follow-up change.

## Migration Plan

1. Land tests and the verifier/stager behind console wiring.
2. Add release-workflow attestation permissions and one attestation per archive.
3. Publish the first attested release; older releases remain downloadable by
   legacy installers but are intentionally ineligible in settings.
4. Bootstrap currently running pre-menu binaries once. Thereafter operators use
   “signed release check and stage” followed by the existing reviewed install.
5. Roll back by reverting console wiring/workflow changes; current executable,
   `.rollback`, and locally staged candidates remain usable.

## Open Questions

None. Automatic installation remains explicitly excluded by the human approval
invariant.
