# Review: a045-add-protection-orders

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Security decision

**ACCEPTED WITH DORMANT SCOPE.** strict attestation schema/parser, pure saga, fake official gateway, migration과 `UNWIRED/OFF` wiring만 승인한다. 실제 conditional mutation, `ProtectionReady=WIRED`, activation 또는 LIVE entry dependency는 사람 attestation 전 금지한다.

## Findings and decisions

1. capability matrix는 account/profile/market/session/type/trigger/quantity/persistence/reservation/idempotency/atomic replace/tool-build/evidence digest를 strict하게 검증한다. legacy/unknown/symlink/owner/mode mismatch는 fail-closed다.
2. first-fill protection은 1초 arm/2초 ACTIVE SLA를 충족하지 못하면 exposure latch와 `PROTECTION_GAP`으로 간다.
3. flatten cancel-confirm deadline은 2초다. 모호하면 `IN_DOUBT`, 최우선 reconcile과 human emergency action으로 전환하고 blind oversell liquidation은 금지한다.
4. older binary가 active protection을 감독할 수 없으므로 binary-only rollback은 금지한다.

## Verification evidence

- OpenSpec strict validation: pass.
- External capability attestation: absent; WIRED/LIVE gate intentionally unavailable.

## Verdict

dormant 범위만 구현할 수 있다. 실제 broker evidence 없이는 이 change의 LIVE capability task와 activation을 완료 처리하지 않는다.

## Dormant implementation record · 2026-07-31

### Pre-Edit Gate

- change/task: `a045-add-protection-orders`, dormant portions only
- target symbols: new strict protection-matrix parser and new pure protection domain/repository
- existing behavior evidence: legacy attest tests/callers, `execgw.ProfileProtection`, official conditional client, engine gateway/reservations, reconciliation and flatten were inspected with current HEAD and refreshed CodeGraph
- upstream inheritance impact: no existing Go function body changed; legacy attestation and UNWIRED entry refusal remain intact
- failing tests first: yes; both new packages failed to compile before implementation and then passed focused/race verification
- safety invariant review: pass for dormant scope; real mutation, WIRED, activation, and operational flatten remain blocked

Function Logic Map: not-applicable — only new Go functions were added. Detailed hard-evidence and branch coverage are recorded in `analysis/dormant-impact.md`.

## Independent security follow-up · 2026-07-31

초기 dormant 구현에 대한 독립 리뷰의 H1-H5/M1-M2 지적을 RED 테스트부터 보강했다.

1. parsing과 verification을 분리했다. parsed matrix는 권한 결과가 아니며, verification은 descriptor와
   정확히 일치하는 외부 evidence bytes 전부를 받아 SHA-256을 다시 계산해야만 성공한다. validity,
   evidence metadata와 capability rows의 canonical matrix digest는 각 evidence descriptor에 함께 결박된다.
2. 파일은 정확한 basename, direct-parent symlink/owner/`0700`, file owner/`0600`, hard-link count 1,
   open identity와 post-read restat가 모두 일치해야 한다. 이는 로컬 integrity 경계이며 signer
   authenticity를 대신하지 않는다.
3. protection account는 separator 없는 8-14 ASCII digits만 허용하며, legacy parser의 임의 문자
   제거 semantics를 재사용하지 않는다.
4. reconciliation 입력과 discrepancy는 account/profile/market/symbol `Scope`를 갖는다. mixed scope와
   duplicate broker ID는 분류 결과를 반환하지 않고 fail-closed한다.
5. saga는 상태별 필드 불변식을 검증하고 `Transition`은 출력도 재검증한다. repository update는 저장된
   row와 비교해 immutable identity, adjacent state transition, monotonic generation/trigger를 transaction
   안에서 확인한 뒤 revision CAS한다.
6. flatten 판단은 start→terminal cancel→sellable observation→deadline 순서, 최대 2초, 동일 scope와
   broker identity, 충분한 quantity를 모두 요구한다. sell claim 합산은 subtraction 방식으로 int64
   overflow를 피한다.

### Remaining explicit blocker

신뢰 signer/signature/trust-root가 명세되지 않았으므로 같은 UID가 작성한 digest-consistent 파일도
authentic attestation으로 간주할 수 없다. 이 follow-up은 dormant parser/domain hardening일 뿐이며
`WIRED`, real gateway, LIVE mutation, engine/UI activation 승인을 추가하지 않는다.

## Independent security re-review · 2026-08-01

- Review scope: `9c42285..46712f4`
- Verdict: **CLEAN FOR DORMANT INTEGRATION**

### Finding closure

- H1: parse/verify 분리, 외부 evidence bytes SHA-256 재검산, canonical matrix digest binding을 확인해 closed.
- H2: account/profile/market/symbol scope 강제와 mixed scope·duplicate broker ID fail-closed를 확인해 closed.
- H3: state-specific saga invariant, transition output 검증, repository identity/state/revision guard를 확인해 closed.
- H4: flatten의 start→cancel→sellable→deadline 순서, 최대 2초, exact scope·broker identity 검증을 확인해 closed.
- H5: sell claim 계산이 subtraction 기반이며 `int64` overflow 경계를 fail-closed함을 확인해 closed.
- M1: exact basename, direct-parent와 file의 symlink/owner/permission, hard-link count, post-read restat 검증을 확인해 closed.
- M2: protection 전용 strict account grammar/canonicalization이 legacy arbitrary removal과 분리됐음을 확인해 closed.

### Dormant boundary and remaining blocker

`execgw.ProfileProtection`은 계속 `UNWIRED`이고 production `cmd/`·`internal/app/` import가 없으며,
real official/trading gateway 또는 broker mutation 구현도 없다. 따라서 이 verdict는 dormant integration
범위에만 유효하고 LIVE 주문, activation 또는 `ProtectionReady=WIRED` 승인이 아니다.

신뢰 signer, signature format, trust-root 배포·회전·폐기 정책은 여전히 미명세다. 이 authenticity
경계가 명세·구현·독립 검증되기 전까지 attestation 결과는 `WIRED` 전환 근거가 될 수 없다.

### Verification

- Focused protection/attestation tests: pass.
- Focused race tests: pass.
- `go vet ./...`: pass.
- OpenSpec strict validation: pass (57/57).

## Signed attestation authenticity implementation · 2026-08-01

### Scope and threat decision

Task 1.1 now uses a strict Ed25519 envelope instead of treating a same-UID digest-consistent file as
authentic. The signed input is domain-separated and length-delimited over envelope version, exact
domain, exact algorithm, key ID and canonical payload. Payload, signature and public key accept only
unpadded canonical base64url; envelope, payload and trust root accept only one canonical strict JSON
encoding, so unknown/duplicate fields and alternate encodings fail closed.

The trust root is a separate-path verifier-only keyset. Immutable startup policy pins its absolute
path, expected owner UID and canonical SHA-256 independently from the runtime attestation. The trust
parent must be real, non-owner traversable and group/other non-writable; the root file is exact
`0444`, regular, single-link and restated after bounded read. Same-parent layouts and same-UID root +
envelope replacement fail closed. There is no runtime writer, TOFU, key download, fallback key,
private signing operation or embedded production public key.

### Key lifecycle and authorization boundary

- unique key ID and public key, exact `PROTECTION_ATTESTATION_SIGNER` role and Ed25519 algorithm;
- not-before/not-after must contain the full matrix issue/expiry window and current verification time;
- overlap rotation accepts either ACTIVE key only during its complete window;
- REVOKED is hard revocation of every signature from that key, including pre-revocation issuance;
- signature success remains parsed/non-authoritative until exact account/profile/market/session/type/
  trigger/quantity, both tool builds and both external evidence byte digests pass.

### TDD and verification

- RED: signed trust/envelope types and verifier were absent; focused package failed to compile.
- GREEN: forged/swap/replay, cross-scope/tool/evidence, unknown/revoked/not-yet/expired keys,
  overlap rotation, duplicate IDs/public keys, wrong role/algorithm/key size, unknown/duplicate/
  noncanonical JSON/base64, separate rootless layout, same-parent/current-UID replacement,
  symlink/hardlink/mode/owner/parent/TOCTOU cases pass.
- `go test ./internal/attest`, focused race and focused vet pass.
- Standard library only; ephemeral private signing helper exists in `_test.go` only.

### Remaining operational blocker

No production keyset, policy digest, signed envelope or human-verified evidence is included. Therefore
runtime provisioning is absent and the system remains `UNWIRED/OFF`; this task does not authorize a
LIVE toggle, broker mutation or `ProtectionReady=WIRED`. Independent Manager security review of the
exact implementation commit is still required before a045 integration.

## Dependency-integrated dormant rebase · 2026-08-01

- Integration base: `70aabdc9936de08df458da13203437ba7d2dd572` (a041/a042 complete).
- Reapplied dormant source range: `9c42285..110fd80`; all three commits replayed without conflict.
- Focused protection/attestation tests, strict OpenSpec validation and whitespace checks pass.
- This rebase does not close the signer/trust-root blocker and does not authorize `WIRED`,
  an operational toggle, or any broker mutation.

## Independent audit BLOCK remediation · 2026-08-01

Independent audit of `1b6da8d` blocked integration on H1-H3 and M1-M4. The implementation was
reworked from RED tests without changing any task checkbox or the dormant execution boundary.

- H1/M1: caller-controlled raw parse/verify/load APIs were removed. `ProtectionVerifier` owns an
  unexported canonical policy source and clock, completes scope/evidence hashing before re-reading
  current policy/root/key and resampling time at final verify, latches monotonic policy generation/
  digest across calls, and returns only the matched authority. A revoke-after-evidence hook test
  proves the final read wins.
- M2-M3: file and direct-parent snapshots plus permission/owner checks run both before and after the
  bounded read; non-Unix ownership support fails closed. Times require exact UTC; evidence,
  capabilities and keys require strict sorted uniqueness; account grammar is 8-14 digits only.
- H2/M4: arbitrary `Repository.Update` was removed. Eight event-specific methods construct private
  events, stored state goes through `Transition`, and one SQL revision CAS is tested through two
  independent SQLite connections for conflicting and same-event races. Last event kind and full
  canonical fingerprint prevent cross-event idempotency collisions. Schema v2 accepts blank legacy
  event identity only for exact revision-1 PLANNED and rejects all other unverifiable rows. Attempt
  and broker lineage mismatches are refused.
- H3: public flatten decision time comes only from the package clock. Success yields an opaque
  scope/broker/quantity/deadline-bound permit with shared atomic one-shot consumption; copied,
  delayed and mismatched permits fail.

Focused tests pass under repeated and race execution, Function Logic Map validation passes, and the
cross-platform ownership split compiles for Windows. This is a remediation record, not a self-issued
security acceptance: independent re-review of the eventual commit remains required. No signer,
trust writer, production verifier constructor, broker gateway, `WIRED`, toggle activation or LIVE
mutation was added.

### Independent remediation re-review

- Verdict: **CLEAN FOR DORMANT INTEGRATION**.
- First pass found and BLOCKed a revoke-during-evidence race and cross-event idempotency collision;
  both were reproduced by RED tests and fixed by final trust sampling plus persisted event identity.
- Second pass found and BLOCKed legacy v1 non-PLANNED/advanced-revision rows with blank lineage;
  migration-time full-row validation and three legacy RED cases closed it.
- Final pass confirmed all three closures and the unchanged `OFF/UNWIRED`, no signer/writer, no
  broker mutation boundary. This is not LIVE/WIRED approval.
