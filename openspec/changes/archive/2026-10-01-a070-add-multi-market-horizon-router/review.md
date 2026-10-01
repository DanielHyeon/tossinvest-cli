# Review — a070-add-multi-market-horizon-router

- Date: 2026-08-03
- Stage: core implementation complete; a067–a069/a066 runtime integration and root gate pending
- Voices: Manager architecture/safety review, independent architecture/test review, round-2 authority-boundary review

## Findings and disposition

- **Accepted blocker:** ownership key is `(account, market, canonical symbol, position_generation)`;
  horizon is admission/attribution only. Routing checks every active horizon owner before scoring.
- Market/horizon rate capabilities are anti-replay/admission subscopes over one physical endpoint and
  reset-generation quota authority; they cannot multiply provider capacity or safety reserve.
- KR and US desired state uses per-market record/revision/lock/activation CAS. Legacy disabled migrates
  both OFF; a verified single-market state may migrate only that market while its peer remains OFF.
- Official exchange calendars and IANA zones define independent sessions; market failure and retry state
  do not cross-contaminate.
- **Resolved blocker:** routing now requires an exact sealed market record/revision in READY state. A caller
  cannot turn an OFF durable market ON by submitting an ON candidate.
- **Resolved blocker:** quota freshness is checked with the authority's trusted clock, not caller-provided
  observation time; backdating cannot revive a stale physical quota snapshot.
- **Resolved blocker:** owner snapshots, ON market records and physical quota snapshots have package-private
  attestation constructors. External callers can inspect values but cannot mint a valid seal. The only public
  market constructor creates a sealed OFF/UNOBSERVED record.
- Rollback accepts OFF targets only. It cannot replay a historical ON activation as fresh authority.
- Inactive same-key owner rows do not compete. Any row from another account/market/symbol/generation corrupts
  the bounded snapshot even when inactive; more than one active same-key row is reconstruction mismatch.

## Verification

- Strict OpenSpec validation: PASS.
- `go test ./internal/strategyrouter -count=1`: PASS.
- `go test -race ./internal/strategyrouter -count=1`: PASS.
- `go vet ./internal/strategyrouter`: PASS.
- `FuzzOwnerKeyNeverIncludesHorizon` and `FuzzLegacyMigrationRetryConverges` (3s each): PASS.
- Cross-horizon ownership, durable OFF binding, shared-quota exhaustion/backdating, concurrent CAS,
  OFF-only rollback and crash migration tests: PASS.
- External-package forgery and exported authority-constructor checks: PASS.
- No combined KR+US approval or automatic activation is introduced.

## Verdict

Core pure/sealed-port implementation approved for integration review. KR and US ship in the same release,
remain independently OFF/UNOBSERVED by default, and share one physical quota authority. Runtime wiring,
broader safety-loop regressions and the repository SDD/gate remain intentionally pending at root integration.

## 처분 감사와 아카이브 (2026-10-01)

- 감사: `analysis/disposition-audit.md`(커밋 `0d3e2849`) — CodeGraph callers 15 · Go AST(`Route` 25 분기 ↔ `RouteSet` 23, 앞 20 원문 동일) ·
  패키지 밖 비시험 `strategyrouter.` 선택자 250 전수. 코어 타입 · owner/market 봉인 · 요청/결정 타입은 a072 · a112 생산 기반으로 생존,
  `Route()` 는 a112 4.3.1 `RouteSet` 이 대체(4.3.2 가드), 스케줄러 durable record · `MigrateLegacy` · `QuotaAuthority` 는 생산 호출자 0.
- Manager 판정: **② 부분 대체 승인**(사용자 거부권 항목으로 보고). ③ 기각 사유 — a112 는 a070 코어 **위에** 서 있다(대체 아님). ① 기각 — 둘째 권한 표면.
- spec delta 는 적용하지 않는다(`--skip-specs`): 7 요구가 정본과 중복 · a112 와 충돌 · 미구현. a123 선례 「구현되지 않은 SHALL 이 정본에 들어가면 안 된다」.
- 미완료 3.1 · 3.4 · 4.2 · 4.3 은 tasks 에 처분을 적고 미체크로 닫는다.
- 후속: quota 단일화는 a112 7.1 설계 입력(`scheduler.BudgetCoordinator` 위에 subscope — Manager 가 a112 소유자에 전달), 섬 코드 삭제는 7.1 뒤
  정리 change 하나(ROADMAP 후속 후보). a112 dependency-matrix 의 a070 행 정정은 a112 소유자 몫.
