## 1. Pre-Edit Evidence and Logic Maps

- [x] 1.1 Run `make sdd-sync`, record CodeGraph definitions/callers/callees/impact for journal migrations, tx-scoped fill apply, Position projection, exit state and reconciliation, and pin the current change base commit.
- [x] 1.2 Complete Go AST artifacts, Function Logic Maps, Branch Test Maps and risk-pattern reports for every existing fill/journal/Position/exit function before editing it, including crash, duplicate and EXIT FIRST branches.
- [x] 1.3 Freeze design D4's complete Campaign/Leg transition tables, prospective-generation CAS, per-order watermark/replacement scheme, command keys, event ordering, stop monotonicity and Position lineage as executable fixtures.

## 2. RED Contract Tests

- [x] 2.1 Add failing domain tests for strategy-neutral campaign/leg identities, ordered sequence, new generation after CLOSED and rejection of embedded 8:4:2, 2:4:8 or seven-leg constants.
- [x] 2.2 Add failing command tests for prospective-generation CAS races, retry idempotence, concurrent expected-version conflict and all-or-nothing first-fill binding crash boundaries.
- [x] 2.3 Add failing transition tests for every Campaign/Leg state-table row, EXIT FIRST races, RECONCILE observation handling, CLOSED terminality and invalid-transition isolation.
- [x] 2.4 Add failing stop tests for saved-stop monotonicity, invalid/missing candidates and source/policy/observed-at provenance.
- [x] 2.5 Add failing per-order watermark tests for duplicate/lower cumulative observations, amend/replacement carry baselines, and replaced/cancelled predecessor late positive deltas; prove retry/restart is delta-zero, commit crash is all-or-nothing, replacement remaining/leg aggregate is recalculated, and cap excess or ambiguous lineage preserves fill/Position while latching campaign RECONCILE and entry block.
- [x] 2.6 Add failing offline reconstruction tests for prospective binding, restart parity, sequence gaps, duplicate keys, orphan order lineage, snapshot drift and the prohibition on automatic repair or broker calls.

## 3. Additive Journal Schema

- [x] 3.1 Add an atomic additive migration for campaign/leg events, prospective-generation tokens, per-order watermarks, projections and nullable lineage with uniqueness, foreign-key and sequence constraints.
- [x] 3.2 Update schema golden tests for migration atomicity, legacy campaign-unknown reads and ErrSchemaTooNew without synthesizing campaign IDs for existing Positions.
- [x] 3.3 Implement journal transaction primitives for prospective position generation CAS, expected campaign version, deterministic command key, per-order watermark, event append and projection update with idempotent result retrieval.

## 4. Campaign Core

- [x] 4.1 Implement strategy-neutral `PositionCampaign`, `CampaignLeg` and order-watermark types plus every row of the frozen Campaign/Leg state-transition tables.
- [x] 4.2 Implement idempotent prospective campaign creation, plan, submit-link, cancel and broker-order cumulative fill commands without broker submission or lane-specific quantities/cadence.
- [x] 4.3 Implement EXIT FIRST admission so EXITING/CLOSING/RECONCILE and unresolved risk-reducing intent reject new exposure while fill detection, reconciliation and emergency exits continue.
- [x] 4.4 Implement long-only effective-stop composition as `max(saved, valid candidate)` with immutable candidate and selection provenance and no protection mutation.
- [x] 4.5 Implement read-only event replay of prospective binding and per-order replacement/watermark lineage plus snapshot comparison that reports stable mismatch reasons and last valid event without changing journal or runtime state.

## 5. Position and Fill Integration

- [x] 5.1 Extend explicit journal lineage from prospective generation through Campaign, Leg and order watermark to decision, intent, mutation attempt, Fill and Position generation while keeping Position quantity/average price as the sole authority.
- [x] 5.2 Integrate first-fill prospective-token binding and per-order watermark/event application into the existing tx-scoped fill hook so fill snapshot, Position delta, exit state and campaign projection commit or rollback together.
- [x] 5.3 Preserve replaced/cancelled predecessor late positive deltas by advancing its immutable watermark and Position exactly once in the fill transaction, recalculating successor remaining and leg aggregates, and latching campaign RECONCILE/new-entry block without truncating over-cap or ambiguous-lineage fills.
- [x] 5.4 Add crash/restart and race integration tests covering concurrent campaign creation, immediate full fill, repeated/replacement partial fill, residual cancel, predecessor late terminal fill, cap excess, ambiguous replacement lineage, exit-versus-scale-in and every RECONCILE recovery row.
- [x] 5.5 Expose only offline reconstruction/read models; leave all production entry callers, live dispatch and lane/automation activation disconnected.
  - 2026-09-07 정정: a065 자체는 그것을 지켰다. 하지만 a072(`8022f578`)가 그 뒤에 배선했다 —
    engine gateway 가 `ApplyPositionCampaignFill` 을 무조건 걸고, a072 first-leg 경로가
    `campaignExposureBlockedInTx` 를 부른다. 이 항목은 **현재 HEAD 에서 더는 참이 아니다**.

## 6. Verify and Gate

- [x] 6.1 Run focused unit, migration, race and journal integration suites and record RED-to-GREEN evidence for every transition-table and Branch Test Map row.
- [x] 6.2 Run broker spies and configuration assertions proving campaign planning/replay emits zero live requests, does not flip lane/automation toggles and never delays stop, emergency exit, reconciliation or fill detection.
- [x] 6.3 Refresh Function Logic Maps, Branch Test Maps and risk reports after edits, then run `openspec validate a065-add-position-campaign-leg-core --strict --no-interactive`, `make sdd-check`, `make test`, `make vet` and `make validate`.
  - 산출물 갱신: `ApplyPositionCampaignFill`(분기 54→49), `Journal.LinkCampaignOrder`(45→46),
    `Journal.CreatePositionCampaign`·`Journal.PlanCampaignLeg`(분기 불변, 파일 해시만 갱신).
    재번호는 손으로 옮기지 않고 옛/새 `ast.json` 을 difflib 으로 정렬해서 했다.
  - 결과는 `review.md` 의 "Gate results" 절에 기록했다.
- [x] 6.4 Complete adversarial independent review for journal atomicity, EXIT FIRST and non-retreating stops, resolve findings and run `make gate CHANGE=a065-add-position-campaign-leg-core` ~~with live entry callers still absent~~.
  - **이 task 의 마지막 조건은 만족할 수 없다.** "live entry callers still absent" 는 a065 가
    쓰일 때는 참이었고 a072 가 배선하면서 거짓이 됐다. 조건을 만족한 척하지 않고 정정한다 —
    측정한 현재 배선은 `issues.md` §0 과 `status.md` 에 있다.
  - 독립 적대 리뷰 네 축(D4/D8 재구성, 원장 원자성, EXIT FIRST·손절 단조성, 휴면 주장 검증)
    전부 최초 판정 BLOCK. P0 셋 + P1 여섯을 해소했고 각 수정은 행동 시험 + 판정이 한 곳인지
    보는 AST 시험 + 수정을 되돌려 RED 를 확인한 뮤테이션을 갖는다(11개 전부 RED, 원복은
    바이트 동일로 검증). 남긴 부채는 `issues.md` §4 에 이름을 붙여 적었다.
  - gate 는 stacked change 규약대로 갈라 돌렸다(memory: 쌓은 change 는 게이트를 깬다):
    change 범위 단계(1–5)는 a065 자기 커밋 슬라이스에서, 저장소 범위 단계(6–11)는 HEAD 에서.
    갈라 돌린 이유와 결과는 `review.md` 에 기록했다.
