# signed-release-distribution Specification

## Purpose
공식 tossctl 릴리스의 공급망 검증, 안전한 추출, 후보 publication과 배포 workflow 계약을 정의한다.

## Requirements

### Requirement: Release archives carry verifiable build provenance
Every platform archive published by the official release workflow SHALL have a
GitHub Actions/Sigstore provenance attestation for its SHA-256. Verification
SHALL use a pinned `sigstore-go` dependency and expected embedded public-good TUF
root digest. It SHALL require successful TUF signature, expiry, consistent
snapshot, version and rollback/freeze validation from a mode-0700 cache, then
successful full Sigstore policy verification of the Fulcio chain/SCT, Rekor
signed inclusion proof/checkpoint, integrated timestamp, certificate validity
at that time, GitHub Actions OIDC issuer, exact official repository
release-workflow identity at the release tag ref, archive subject digest, and
SLSA provenance v1. The predicate SHALL bind the official source
repository/workflow/tag ref and resolved commit; the extracted binary SHALL
carry that VCS revision and SHALL NOT be marked modified. Merely finding those
fields or matching a sibling checksum SHALL NOT satisfy verification.

#### Scenario: Official attested archive
- **WHEN** an archive digest was attested by the official release workflow at the matching stable tag
- **THEN** provenance verification succeeds and reports the bound digest and workflow identity

#### Scenario: Archive and checksum are replaced together
- **WHEN** a downloaded archive and its sibling checksum agree but no valid official provenance binds that digest
- **THEN** verification fails and no executable bytes are staged

#### Scenario: Different workflow or ref signed the archive
- **WHEN** a valid Sigstore bundle identifies another repository, workflow, branch, or tag
- **THEN** verification fails even though its cryptographic signature is otherwise valid

#### Scenario: TUF metadata is expired, rolled back, frozen, or corrupt
- **WHEN** the trust metadata cannot pass the pinned-root TUF client's freshness and rollback checks
- **THEN** verification fails closed and no cached or downloaded bundle is accepted

#### Scenario: Rekor evidence is only present but not valid
- **WHEN** a bundle carries integrated-time or inclusion fields whose signatures, checkpoint, certificate time, or inclusion proof do not verify
- **THEN** verification fails and no candidate is staged

#### Scenario: Candidate commit differs from provenance
- **WHEN** an archive is signed for the tag but its extracted Go binary reports another VCS revision or a modified tree
- **THEN** staging fails

### Requirement: Latest release discovery is fixed and bounded
The release client SHALL query only its constructor-bound official repository
and latest-stable endpoint, reject drafts, prereleases, non-canonical stable
tags, and unsupported platforms, and select only the exact
`tossctl-<goos>-<goarch>` archive name. All requests and redirects SHALL remain
HTTPS, reject userinfo/non-443 ports, use the exact bounded
release/attestation/TUF host rules, have redirect/page/bundle/time/byte limits,
and reject non-success responses. `Content-Length`, when present, SHALL be an
early limit check while bounded reads enforce the limit for chunked responses.
Public requests SHALL send no Authorization header. Caller-supplied URL,
repository, tag, asset, destination, or signer values SHALL NOT affect
production selection.

TUF retrieval SHALL use an injected context-aware fetcher that accepts only the
exact public-good CDN host over HTTPS/default port, follows no redirect, applies
caller cancellation and an explicit HTTP timeout, and enforces an independent
response cap in addition to TUF metadata lengths.

For a current canonical release version the latest tag SHALL be strictly newer.
For a non-release development build only, staging a signed stable release MAY
act as bootstrap and SHALL report that version ordering was unavailable.

#### Scenario: Latest stable platform asset exists
- **WHEN** the official latest release is stable and contains the exact current-platform archive
- **THEN** the client downloads only that archive and requests attestations for its locally computed SHA-256

#### Scenario: Redirect escapes the trusted host set
- **WHEN** metadata, archive, attestation, or bundle delivery redirects to an unapproved host or non-HTTPS scheme
- **THEN** the request fails before response bytes can be accepted

#### Scenario: TUF request stalls or redirects
- **WHEN** a TUF request redirects, exceeds its response cap, reaches its client timeout, or its caller is cancelled
- **THEN** trust refresh fails promptly and no release operation or candidate publication continues

#### Scenario: Response exceeds its limit
- **WHEN** any metadata, bundle, archive, or extracted binary exceeds its configured maximum
- **THEN** staging fails and the existing candidate remains unchanged

#### Scenario: Valid historical release is replayed
- **WHEN** the latest response identifies a tag equal to or older than the running canonical release
- **THEN** staging refuses the downgrade even if its attestation is valid

#### Scenario: Development build bootstraps
- **WHEN** the running binary has no canonical release version and the latest stable archive otherwise passes every verification
- **THEN** it may be staged with an explicit bootstrap warning and without claiming an ordered upgrade

#### Scenario: Attestation list is paginated
- **WHEN** GitHub returns multiple bounded pages or bundle URLs for the digest
- **THEN** the client follows only bounded allowlisted cursors and accepts only a bundle that independently passes the complete policy

#### Scenario: Bundle URL names another Azure tenant
- **WHEN** GitHub returns a bundle URL outside exact host `tmaproduction.blob.core.windows.net`, outside `/attestations/`, or requiring a redirect
- **THEN** bundle retrieval fails closed before accepting response bytes

### Requirement: Verified archive extraction cannot select a filesystem path
The downloader SHALL verify the complete archive before extraction. It SHALL
disable gzip multistream input; cap compressed bytes, aggregate uncompressed
bytes, entry/header count and binary bytes; read entries without materializing
their paths; accept exactly one regular root entry named `tossctl` on supported
Unix systems; and reject duplicate names, PAX/GNU/sparse path overrides,
NUL/backslash/absolute/`..` paths, links, devices, FIFOs, and oversized content.
Only zero-size directory headers and bounded regular files at `auth-helper` or
below `auth-helper/` may be ignored. It SHALL pass only the extracted binary
bytes to the fixed candidate publisher. Candidate code
SHALL NOT execute during download, verification, extraction, or publication.

#### Scenario: Archive contains traversal and a binary
- **WHEN** a verified archive also contains an entry whose path escapes the archive root
- **THEN** extraction rejects the archive and writes no archive-selected path

#### Scenario: Archive contains a linked tossctl
- **WHEN** the exact binary entry is a symlink or hard link
- **THEN** extraction rejects it and no candidate code executes

#### Scenario: Archive contains one valid tossctl
- **WHEN** a verified archive contains one bounded regular `tossctl` entry
- **THEN** only its bytes are submitted to fixed candidate validation

#### Scenario: Official archive contains auth-helper directories
- **WHEN** a verified archive contains zero-size directory headers and bounded regular files under `auth-helper/`
- **THEN** those entries are safely consumed without writing paths and the exact root `tossctl` remains the only staged content

#### Scenario: Archive expands beyond aggregate limits
- **WHEN** compressed input, decompressed entries, header count, gzip members, or sparse/PAX metadata exceeds its bound
- **THEN** extraction fails without publishing any bytes

### Requirement: Candidate publication is atomic and preserves prior state
The fixed candidate publisher SHALL serialize inspection, staging, and install;
write and sync a same-directory temporary regular executable; validate SHA-256,
Go build information, module, exact `cmd/tossctl` command identity, VCS
revision/modified state, and GOOS/GOARCH without execution; and then atomically
publish it to `<running-path>.candidate`. It SHALL sync a no-follow regular-file
recovery copy before replacement. A failure before replacement SHALL leave the
previous candidate; a later failure SHALL restore it and sync, or retain and
report an explicit recovery path if restoration itself fails. It SHALL NOT
claim power-loss durability after a failed fsync. The normal installer SHALL
still re-open and revalidate the candidate against the operator-reviewed SHA
before replacing the running executable.

#### Scenario: Extracted executable has the wrong command identity
- **WHEN** verified archive bytes contain a Go executable from the repository but not `cmd/tossctl`
- **THEN** publication fails and the prior candidate remains

#### Scenario: Publication succeeds
- **WHEN** verified extracted bytes pass fixed candidate validation and directory sync
- **THEN** the sibling candidate atomically becomes those bytes and is available for separate human review

#### Scenario: Candidate directory sync fails
- **WHEN** publishing the new candidate cannot be durably synced
- **THEN** the prior candidate is restored and synced, or an explicit retained recovery path and both failures are reported

#### Scenario: Stage and install race
- **WHEN** staging, inspection, and installation target the same updater concurrently
- **THEN** updater-level serialization prevents mixed candidate bytes and each completed operation observes one whole candidate state

### Requirement: Official release workflow is least-privilege and immutable
The official release workflow SHALL pin every third-party action to a full commit
SHA, default permissions to none, grant build jobs only contents-read,
id-token-write, attestations-write, and the artifact-metadata-write permission
required by the pinned attestation action, and grant release publishing
contents-write only to the publishing job. It SHALL build only a canonical tag,
attest each final archive before release publishing, and SHALL NOT replace,
clobber, or edit assets of an existing attested release. A changed build SHALL
require a new tag.

#### Scenario: Release workflow is rerun after publication
- **WHEN** the tag already has a published attested release
- **THEN** the workflow refuses to replace or edit its assets
