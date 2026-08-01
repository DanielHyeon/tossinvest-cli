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

## Full dormant core implementation candidate · 2026-08-01

The candidate now adds the official-only adapter and complete controller behavior behind an
unmintable `Activation`. Create/replace/cancel attempts are durable before dispatch; response loss,
missing identity, unknown lifecycle, cancel disappearance, mismatched readback, pagination overflow,
and deadline misses all latch entry closed and return `IN_DOUBT`. Recovery reads bounded OPEN and
CLOSED official pages and correlates exact broker/client identity. The a041 bridge accepts only its
typed immutable snapshot and one-way trigger movement. Flatten remains cancel-first and yields only
a two-second exact-scope one-shot authorization after authoritative sellable quantity.

The StockOS-style `exit-protection` lane console is status-first and defaults to OFF/UNWIRED. It has
no text, number, textarea, select, contenteditable, free symbol/trigger/quantity/reason, typed
confirmation, LIVE toggle, or enable-all control. The only submitted values are hidden CSRF/opaque
server capabilities and the required weakening checkbox; the server enforces the three-second delay.

This is an implementation candidate awaiting independent review and final gate, not self-approval.
No production constructor/minter/provisioning/app wiring was added, no toggle was changed, and no
LIVE order was sent. A later operational change must separately provision signed evidence, preserve
OFF by default, pass the full gate, record explicit operator approval, and prove recovery/rollback
readiness before any activation.

## Independent review BLOCK remediation candidate · 2026-08-01

Review of `4679fe2` found that a fresh controller trusted durable ACTIVE rows before a current broker
read, some error paths could leave the entry latch open, broker calls inherited potentially unbounded
contexts, and create/cancel/recovery identity checks were not exact enough. RED tests now cover
restart, concurrent registration, missing and terminal-without-trigger recovery, body and database
errors, blocked create/list/cancel calls, exact mutation receipt/readback, duplicate client identity,
and cancel-first authorization under race.

The controller now starts closed whenever any saga exists and opens only when every scoped saga is
ACTIVE and has an exact broker confirmation from this controller instance. Every fill/register/
replace/reconcile/recover/flatten operation closes the latch before validation or I/O; error,
timeout, unknown, terminal-without-trigger, disappearance, and failed durable recovery paths keep it
closed. Terminal-without-trigger and disappearance are typed `ErrProtectionGone` and are durably
projected to RECONCILE where the prior state permits it. CLOSED/other non-ACTIVE rows never satisfy
the latch.

Register carries the remaining two-second ACTIVE budget after the one-second arm admission check;
replace/reconcile/recover use a two-second broker context, and cancel plus sellable observation share
the flatten operation's original two-second budget. Parent cancellation/deadlines propagate and a
callee returning after cancellation is still treated as `IN_DOUBT`.

The official adapter requires the mutation receipt client identity, requested broker ID equal to the
detail row ID, and exact client ID, account-scoped gateway binding, profile/market/symbol scope,
SELL/SINGLE/MARKET/STOP shape, integer trigger/quantity and expiry on create/replace confirmation.
Cancel and recovery use an exact durable `BrokerTarget`; a mismatched or duplicate broker/client row,
unknown lifecycle, missing row, or 404 stays ambiguous. The official detail schema does not expose an
account-ref field, so account identity is established by the pre-bound account-scoped official client
plus equality between the target scope and immutable gateway scope; it is never inferred from symbol.

This remediation does not add a production activation minter, app/cmd construction, toggle, or free
UI input, and does not authorize broker execution. It remains pending independent re-review and gate.

## Restart inventory and replace-lineage BLOCK remediation · 2026-08-01

The public controller constructor now always starts with entry closed. An empty local repository is
not proof of an empty broker account; only a bounded authoritative OPEN+CLOSED inventory read can
open the latch. Empty inventory reconciliation is covered explicitly, while broker read failure,
timeout, ambiguous identity, any non-ACTIVE durable saga, and any orphan keep it closed.

CREATE and REPLACE canonical bodies plus acknowledged target/result broker IDs now form the durable
identity chain. Recovery of a response-lost replacement selects the unique non-terminal row matching
the pending canonical body, records its result ID before activating the saga, and thereafter walks
the acknowledged chain from the current result to its predecessors. A historical row is ignored
only when its broker ID, client ID, trigger, quantity, expiry, shape, terminal state, and untriggered
state all exactly match that chain. Missing links, cycles, duplicate current/retired rows, unrelated
same-client rows, identity mismatch, and orphan live rows fail closed as `IN_DOUBT`.

The official cancel fallback applies the same typed predecessor list across bounded OPEN+CLOSED
pages; it never treats a client ID match alone as the current order. Tests cover two successive
replacements, restart recovery and reconciliation with old CLOSED plus new OPEN rows, response loss,
repeated terminal cancel observation, and unrelated historical identity. Focused repeated tests,
focused race tests, the full Go suite, `go vet`, whitespace checks, and Function Logic Map validation
pass. This remains a remediation candidate awaiting independent re-review; OFF/UNWIRED and the
input-free StockOS-style console are unchanged.

## Mutation commit-window and exact-expiry BLOCK remediation · 2026-08-01

Recovery now interprets the saga's durable CREATE/REPLACE attempt across every pre-commit state.
PLANNED is never claimed as broker-dispatched. DISPATCHED and IN_DOUBT treat the result as unknown
and require exactly one bounded inventory row matching the canonical request identity; zero or two
or more candidates remain `IN_DOUBT`. ACKNOWLEDGED requires its durable result broker ID and resumes
the interrupted saga commit against that exact row. For replacement, the acknowledged new result is
the current generation and the target plus its checked predecessors are retired; generation gaps,
forks, target mismatch, body mismatch, and result mismatch fail closed.

When recovery uniquely resolves an unknown dispatch, it writes DISPATCHED/IN_DOUBT to ACKNOWLEDGED
with the selected result ID before applying ACTIVE to the saga. A second crash in that interval is
therefore handled by the ACK path. RED/GREEN tests cover CREATE and a repeated REPLACE chain at both
DISPATCHED-to-ACK and ACK-to-saga-commit windows, followed by a fresh restart reconciliation. They
also prove PLANNED cannot claim a coincidental row and unknown inventory cardinality 0/2 cannot open
entry.

Expiry is now part of the mandatory current `BrokerTarget`, not an optional comparison. Empty or
non-canonical target expiry is rejected; official target matching has no empty-expiry wildcard.
Create and replace response confirmation, cancel identity, restart recovery, and reconciliation all
require the exact canonical expiry. The detail-available cancel path verifies that identity before
issuing DELETE. Expiry mismatch tests cover response confirmation, Get/Cancel, and restart inventory.

This follow-up changes no UI, activation minter, app construction, execution profile, toggle, or
transport boundary. The controller still starts closed and all broker reads remain bounded. It is a
review candidate, not LIVE/WIRED approval.

## Durable SLA clock and mandatory cancel preflight BLOCK remediation · 2026-08-01

Registration no longer derives its one-second arm or two-second ACTIVE deadline from caller time.
The PLANNED saga's persisted `UpdatedAt` is the durable fill timestamp before registration begins;
the retained compatibility argument must match it exactly, and every deadline/EvaluateArm decision
uses the persisted value. A delayed caller cannot pass `now` to restart either SLA, while a nearby
but non-identical timestamp is also rejected before broker dispatch.

Flatten timing now starts from the controller's internal clock at operation entry. The legacy caller
argument is ignored, and one absolute internal `start + 2s` deadline clamps the shared cancel and
sellable context as well as the final decision. Future caller time cannot extend a slow operation;
past caller time cannot force a false early deadline. Tests verify both cases and an actual broker
context deadline no later than the internal two-second budget.

Official cancel now requires a successful exact `ConditionalOrderRaw` preflight before DELETE.
404, timeout, transport error, unknown lifecycle, body/identity mismatch, and expiry mismatch all
return ambiguous with zero cancel calls. After a successful preflight, DELETE still requires the
existing authoritative terminal non-triggered observation; post-delete disappearance remains
`IN_DOUBT` rather than assumed cancelled. These changes retain constructor-closed behavior, bounded
recovery, exact mutation lineage, OFF/UNWIRED, and the input-free StockOS-style console.

## Complete lifecycle pagination BLOCK remediation · 2026-08-01

Post-delete cancel fallback now requires independent authoritative completion of both OPEN and CLOSED
conditional-order scans before a terminal observation can authorize flatten. Finding the exact target
early is not sufficient. Each lifecycle tracks visited cursors and accepts only an explicit terminal
page with `HasNext=false` and an empty next cursor. Ten-page exhaustion, `HasNext=true` with an empty
cursor, repeated/non-progressing cursor, cursor cycle, contradictory pagination metadata, or any list
error returns ambiguous even when an exact terminal target was already found.

RED/GREEN coverage places the target on the first page with a possible duplicate on page eleven,
exercises empty, repeated, and cyclic cursors, proves OPEN-complete/CLOSED-incomplete and the inverse
both fail, and retains an exact success case only when both lifecycle scans terminate completely.
Controller cancel errors still produce no flatten authorization. Mandatory exact preflight,
post-delete disappearance handling, internal deadlines, durable mutation lineage, OFF/UNWIRED, and
the input-free console remain unchanged.

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
