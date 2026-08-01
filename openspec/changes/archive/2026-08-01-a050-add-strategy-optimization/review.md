# Review: a050-add-strategy-optimization

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability, UI/UX

## Findings and decisions

1. StockOS 최신 lane-console의 화면 단위 navigation, changed-key-only save와 effective refetch를 참고한다. 구형 slider matrix는 복제하지 않는다.
2. 모든 조작은 stable preset/radio/select/chip/toggle/discrete-step/current-row action이다. text/number/textarea/contenteditable/free symbol/typed phrase/free reason은 0개다.
3. a041 `internal/settingmeta`를 조합하며 owner descriptor가 유일한 label/default/help source다. a050은 category/lifecycle만 소유한다.
4. console/httpapi는 journal read-only이고 durable commander를 공유한다. engine만 trading journal state를 쓴다.
5. high-risk apply/rollback은 desired를 보존할 수 있으나 manifest가 바뀌면 effective entry는 OFF다. rollback은 current constraints를 검증한 새 version이다.
6. evidence 부족은 자동 추천을 막지만 검증된 보수적 server preset의 human selection 자체를 막지 않는다.

## Verification evidence

- OpenSpec strict validation: pass.
- StockOS reference inspected: latest lane-console shell/tabs/human-only control; old full-page slider matrix rejected.
- Registry coverage is frozen to the sole approved a041 `exit.common-policy`
  `settingmeta` descriptor. a044/a046/a047 remain owner-specific read-only
  projections; a050 does not invent writable metadata or activation authority.
- a049 evidence is opened only through an existing-checkpoint, immutable,
  query-only adapter. Missing DB, active WAL/SHM, file drift, query mismatch,
  missing lineage/metrics, fewer than 20 samples, or source age over 72 hours
  fails closed.
- Candidate capabilities HMAC-bind candidate ID, actor, base/category/origin,
  raw changes/evidence, timing and derived risk state. Apply revalidates the
  current registry, snapshot, actor, evidence digest/status and CAS version.
- Schema v3 migration verifies legacy digests before expanding them, protects
  every immutable snapshot field, binds the mutable control pointer with a
  digest, and rejects drifted/no-op append-only trigger definitions. Apply
  validates that pointer before either a fresh write or an idempotent replay.
- Every audit event carries its own digest. A bounded history read first checks
  whole-ledger snapshot/application/candidate/audit coverage, so deleting all or
  part of a multi-change audit trail fails closed instead of disappearing below
  the newest-1000 display limit.
- Optimization mutations accept only an exact URL-encoded field set. Multipart,
  missing, duplicate, unexpected and client-invented values are refused before
  the commander.
- UI verification covers 360px/420px containment, 44px targets, keyboard focus,
  CSP, error/stale/insufficient states, three-second risk review, and exactly
  zero free text/number/range/textarea/contenteditable/symbol/reason inputs.
- Focused and full `go test ./... -count=1`, focused race, vet, Windows compile,
  and `git diff --check` pass.

## Independent review resolution

The first adversarial security/integration pass blocked on writable performance
DB construction, unbound candidate actors, multipart ambiguity, incomplete
snapshot integrity, and laundered evidence freshness. A later test-architecture
pass also found control-pointer rollback through idempotent replay and deleted
audit-row invisibility. Each item received a RED test and fail-closed
implementation. The final independent security re-audit reports 0
critical/high/medium/low findings, and the final independent test/maintainability
re-review reports PASS with no blocker. Its non-blocking observations—large
`store.go` size and no expired-candidate retention policy—are recorded for later
maintenance rather than weakening this change's integrity contract.

Accepted boundaries are explicit: the console currently records the shared
role `console-operator`; per-session identities are supplied later through
`Store.ForActor`. Same-UID administrator tampering is outside the local DB
integrity model. Engine activation/manifest consumption, LIVE, gate, kill switch,
and active-position mutation remain out of scope and have no a050 route.

## Verdict

Implementation review PASS, subject only to the recorded `make gate
CHANGE=a050-add-strategy-optimization` landing gate. LIVE/activation authority
remains outside this change.
