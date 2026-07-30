# Proposal-freeze review

Date: 2026-07-30

Initial verdict: **NO-GO**

Independent adversarial review covered engineering, supply-chain security,
operator UX, and delivery risk. Implementation was held until the following
findings were resolved in the design and delta specifications.

## Resolved P0 findings

- TUF bootstrap is now the pinned `sigstore-go` embedded public-good root plus
  an expected root digest. The contract requires TUF signature, expiry,
  consistent-snapshot, version, rollback, and freeze checks and a mode-0700
  cache. Sigstore acceptance requires complete library policy success, not JSON
  field presence: Fulcio/SCT, Rekor signed inclusion/checkpoint, integrated-time
  certificate validity, exact issuer/SAN, predicate, and digest.
- The release workflow is now part of the trust boundary. All actions must be
  pinned by full commit SHA, permissions are job-minimal, canonical tag builds
  are enforced, final archives are attested before release, and existing release
  assets cannot be clobbered or edited.

## Resolved P1 findings

- Canonical installed releases reject equal/older tags. A development build has
  one explicit bootstrap exception and the UI cannot call it an ordered upgrade.
  Provenance source ref/commit and candidate VCS revision must agree.
- Redirects reject userinfo, non-443 ports, literal/private hosts, suffix tricks,
  and more than five hops. GitHub attestation pagination, bundle count, pages,
  and aggregate bytes are bounded; bundle retrieval is fixed to
  `tmaproduction.blob.core.windows.net/attestations/` with no redirect.
- Archive limits now cover compressed bytes, aggregate uncompressed bytes,
  headers, multistream gzip, PAX/GNU/sparse metadata, traversal spellings,
  duplicates, and all non-regular special entries.
- Candidate publication now defines updater-level serialization, recovery-copy
  and fsync order, restoration behavior, and the honest retained-recovery result
  when restoration itself fails.
- The verification plan now includes real publisher/install/download races under
  the race detector in addition to route fakes.
- The checksum-only `tossctl update` path remains compatible but must warn that
  it is deprecated and direct operators to the signed settings flow. Replacing
  it is an explicit follow-up, not hidden equivalence.

## Resolved P2 findings

- Missing `Content-Length` is accepted; bounded reads are authoritative.
- Restarted consoles label an existing candidate's provenance unknown until
  signed download verification succeeds again.
- Errors distinguish discovery/rate-limit, trust bootstrap, attestation,
  extraction, candidate validation, and restoration failures.

The first re-review found three remaining blockers: reverse-completion rollback
between concurrent release checks, an over-broad Azure tenant suffix, and
unspecified `auth-helper/` directory headers. The final contract adds a
release-only whole-operation single-flight mutex, exact production bundle
host/path with no redirects, and safe zero-size directory handling plus the
requested race/fixture tests.

Final independent freeze verdict after contract changes: **GO**. Strict OpenSpec
validation passes; implementation remains subject to the
RED/GREEN/REFACTOR/VERIFY gates in `tasks.md`.

## Independent implementation review

Date: 2026-07-30

Initial implementation verdict: **NO-GO**

The reviewer found that `tuf.DefaultOptions().WithContext(ctx)` did not make
go-tuf's default synchronous fetcher use that context. Its default HTTP client
also had no caller-controlled timeout or redirect boundary, so a TUF request
could leave the fixed host set or hold the release single-flight indefinitely.

The implementation now injects a custom fetcher with `WithFetcher`. It validates
the exact public-good TUF CDN HTTPS host before I/O, rejects userinfo,
non-default ports and IP literals, follows no redirect, creates every request
with the caller context, caps the HTTP client at 30 seconds, and enforces both
TUF's supplied maximum and an independent 16 MiB bounded read. Regression tests
cover the production URL policy, redirects, chunked overflow, caller
cancellation and HTTP timeout.

Final independent implementation verdict after remediation: **GO**. The
reviewer confirmed focused and race tests, strict OpenSpec validation, and
`git diff --check`; no P0/P1/P2 blocker remains.
