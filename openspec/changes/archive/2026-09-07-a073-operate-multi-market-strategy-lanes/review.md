# Review — a073-operate-multi-market-strategy-lanes

- Date: 2026-08-04
- Stage: complete exact-digest dormant deployment after reviewed fix-forward recovery
- Voices: Manager scope review, independent operations/security/UI review, projection contract review

## Findings and disposition

- Console, JSON and SSE consume one server-owned market projection. KR and US have separate status/error
  envelopes, exact `WIRED|UNWIRED` readiness and honest unavailable/not-configured states.
- The existing authenticated GET-only strategy-runtime pattern is reused; no order, gate, activation,
  autostart, protection-weakening route or free input is added.
- Performance attribution requires the complete persisted market-to-close identity chain and conserves
  partial/staged-close quantity, basis, fees, taxes, FX and PnL. Missing facts are `link_missing` or
  `not_measured`, never zero/current-FX inference.
- Compose replacement pins image/schema/config/activation/protection/volume preimages, replaces one service
  at a time and reverse-rolls only the replaced subset. When complete compatibility evidence proves entry
  remains OFF, incompatible rollback retains that observed state and forbids destructive downgrade.
- Deployment actions seal UTC issue/deadline times within five minutes. Their canonical evidence digest binds
  action/service/image/schema/health/state/environment/mount facts, so replay, late observations and mutation
  after evidence generation are rejected without advancing the plan.
- Independent HIGH review found three fail-closed gaps: an applied-but-unhealthy current attempt could be
  omitted from rollback, arbitrary/unbounded health digests could be replayed, and an unhealthy compatibility
  read could authorize rollback. The implementation now distinguishes `APPLIED|NOT_APPLIED`, rolls the applied
  current attempt first, seals the action window/canonical observation and refuses destructive rollback on
  unhealthy, timed-out or invalid compatibility evidence.
- Follow-up HIGH review found recovery records could falsely report entry `OFF` during state drift and health
  failure could hide simultaneous preservation drift. Recovery now reports exact common observed `ON|OFF`
  only from complete KR/US evidence, otherwise `UNKNOWN`; preservation drift is classified before schema or
  health failure and missing compatibility evidence cannot authorize rollback.
- The implemented shared model owns the exact paired KR/US shape and validation. Console, REST, SSE and the
  authenticated Unix reader consume that model without recomputing readiness, effective state or refusal.
- Unix transport review confirms a bounded strict decoder, exact private descriptor/socket permissions,
  no-follow plus same-file descriptor checks, constant-time bearer comparison and a `Read`-only client type.
- Integrated partial-failure and reconnect tests preserve the unaffected market byte-for-value and replace a
  prior partial state with one complete fresh snapshot; no zero/current/cross-market fallback exists.
- The new performance view is an immutable leaf over supplied authoritative evidence. Exact market/lane/
  version/campaign/leg and decision-to-close identity prevents same-ticker or cross-scope laundering; exact
  replays deduplicate while divergent replay and correction overrun fail closed.
- Partial entries, staged closes and corrections conserve quantity and authoritative cost basis. Source and
  reporting PnL expose fee/tax/FX provenance and policy-versioned rounding; missing close, fee or persisted FX
  evidence is `not_measured`, including source==reporting currency, rather than an invented zero/rate-one.

## Design and DX disposition

The plan extends the existing `/strategy-runtime` information architecture rather than introducing a new
interaction model. Primary scan order is market identity → desired/effective/refusal → evidence/scheduler/
protection/reconciliation → campaign/risk → provenance/performance. Loading/unavailable, dormant, partial,
current and lineage-missing states are specified; mobile/accessibility and console/API parity reuse existing
golden-contract patterns. Operator time-to-diagnosis improves because every blocked market exposes its first
typed refusal without requiring journal joins.

## Projection-wave verification

- Strict OpenSpec validation: PASS.
- `go test -count=1 ./internal/strategyprojection ./internal/strategyprojectionrpc ./internal/console ./internal/httpapi`: PASS.
- The same four-package command with `-race`: PASS.
- `go vet` for the same four packages: PASS.
- Targeted OpenAPI/strategy-runtime contract tests: PASS.
- `git diff --check`: PASS.
- Legacy single-market console coverage was replaced by paired authority-projection, invalid/read-error
  fail-closed, authenticated GET/HEAD, no-input/no-mutation, responsive/CSP, partial-market and real Unix
  console/API/SSE convergence tests.
- Full repository gates, real Compose preimage verification and final implementation review passed.

## Lane-performance verification

- `go test -count=1 ./internal/performance`: PASS, including the existing million-row bounded query fixture.
- `go test -race -count=1 ./internal/performance` and `go vet ./internal/performance`: PASS.
- The implementation adds new performance-only leaf functions and tests; it does not change the existing DB
  schema, pruning functions, journal adapter, execution path or any operating authority.

## Deployment-guard verification

- Pure `internal/deployguard` package and repository boundary tests cover immutable preimage refusal, exact
  rendered target images, frozen service order, bounded sequential actions, canonical evidence, typed
  timeouts, applied/not-applied subset accounting, reverse rollback and incompatible/read-failed recovery.
- `go test -count=1`, `go test -race -count=1`, `go test -count=25` and `go vet` for
  `./internal/deployguard`: PASS. Existing Compose/API separation static test, strict a073 OpenSpec validation
  and `git diff --check`: PASS.
- The human-authorized dormant deployment used exact immutable images and one-service-at-a-time replacement.
  A first engine startup failure stopped further rollout; an exact-image rollback was attempted, refused by
  the old binary's `v19` ceiling after migration to `v29`, and correctly converted to typed fix-forward
  recovery with entry OFF. The corrected commit then passed independent review and the a072 gate before both
  services were replaced and verified healthy.

## Verdict

The paired KR/US operational projection, lane performance, deployment guard and exact-digest dormant release
are CLEAN. Existing safety loops are running and both markets remain independently OFF/NOT_CONFIGURED with
zero strategy/protection mutation rows. No LIVE order, operating toggle, approval or market activation was
performed. Market activation remains an explicit later human decision.

## 완료 게이트 (2026-09-08, 아카이브 시점)

이 절은 change 를 닫으면서 **실제로 돌린 것**만 적는다.

| 단계 | 결과 |
|---|---|
| 1 tasks.md 존재 | OK |
| 2 미완료 태스크 | 0건 (27/27) |
| 3 deploy-pair.txt | 선언 없음 — 단독 배포 |
| 4 review.md | 이 파일 |
| 5 Function Logic Map | **a073 자기 완료 커밋(`fb135d85`)에서 PASS** — `evidence complete or diff-proven exempt` |
| 6 make sdd-check | RC=0 |
| 7 make test | RC=0 |
| 8 make test-seams | RC=0 |
| 9 make test-race | RC=0 |
| 10 make vet | RC=0 |
| 11 make validate | RC=0 (61/61) |

### 5단계를 왜 HEAD 가 아니라 자기 커밋에서 쟀는가

`check_analysis` 는 change 의 base commit 과 **워크트리**를 비교한다. a073 의 base 는
`171739a4` (2026-08-04) 이고 HEAD 는 그로부터 한 달 넘게 앞서 있으므로, HEAD 에서 재면
a074~a121 이 바꾼 함수 전부가 "a073 이 증거를 안 냈다"로 나온다. 그래서 두 자리에서
재고 차이를 귀속했다.

| 측정 | 지적 수 |
|---|---|
| HEAD (수리 전) | 338 |
| a073 자기 완료 커밋 `fb135d85` (수리 전) | **2** |
| a073 자기 완료 커밋 (수리 후) | **0** |
| HEAD (수리 후) | 336 |

즉 a073 에게 귀속되는 것은 2건이고 나머지 336 은 쌓임 아티팩트다. 그 2건은 아래에서
고쳤다.

### 고친 2건 — 이름이 바뀐 테스트를 가리키던 인용

a073 의 커밋 `8022f578` 이 테스트 함수 둘의 **이름과 본문**을 바꿨다.

- `TestCampaignCoreHasNoProductionBrokerOrToggleWiring` → `TestCampaignCoreProductionWiringHasNoBrokerOrToggleAuthority`
- `TestTheApplyAnswerNamesTheMarketTheCurrencyCloses` → `TestTheApplyAnswerNamesPairedAccountBaseFXRequirements`

이 change 는 FLM 증거를 a072 에서 빌려 쓰는데(`analysis/function-logic-reference.txt`),
그 두 번들의 branch-test-map 이 **옛 이름**을 덮는 테스트로 계속 인용하고 있었다. 인용된
이름은 현재 트리 어디에도 없으므로 그 커버리지 주장은 아무도 답하지 않는 주장이었다.

당시 게이트는 이것을 못 봤다 — 인용한 테스트가 실재하는지 보는 검사는 a073 보다 **나중에**
생겼다. 그러므로 이것은 그때 한 거짓말이 아니라, 더 엄격해진 검사가 드러낸 낡은 인용이다.

수리는 인용만 현재 이름으로 옮겼다. `ast.json` 은 `revision: base` 라 base 시점 이름과
좌표를 그대로 둔다 — 그것이 그 번들이 기술하는 대상이다. 옛 이름은 backtick 없이 적었다.
backtick 을 두르면 그것도 "지금 트리에 있어야 하는 테스트"라는 인용이 되어 방금 고친
거짓을 다시 만들기 때문이다.

### 게이트 도구 결함 하나 — 아카이브가 참조를 고아로 만든다

5단계를 처음 돌렸을 때 나온 것은 지적 목록이 아니라
`function-logic reference base is invalid: missing base-commit.txt` 였다.

원인은 a073 이 아니라 해결자다. `check_analysis.py` 는 참조된 change 를
`openspec/changes/<id>` 에서만 찾았고, a072 는 2026-08-29 에 아카이브되어 그 자리에
없었다. **증거를 빌려주는 쪽이 먼저 아카이브되면 빌리는 쪽의 게이트가 영구히 막힌다.**

증상을 우회(번들 231개 복사·포인터 재작성)하지 않고 해결자를 고쳤다:
`archive/<YYYY-MM-DD>-<id>` 도 본다. 날짜 접두사를 벗긴 나머지를 **전부** 맞추고
(접미사 일치는 남의 증거를 통과시킨다), 같은 id 의 아카이브가 둘이면 고르지 않고 멈춘다.
시험 넷과 반증 셋은 `tools/logic-map/test_check_analysis.py` 에 있다.
