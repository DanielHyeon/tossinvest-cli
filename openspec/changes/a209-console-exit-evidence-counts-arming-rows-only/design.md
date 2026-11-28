# a209 design — 청산 근거 연결은 무장 행만 센다

## 1. 사실(현재 HEAD `baeddfe31`, 분기는 AST 열거 기준)

분기 근거: `analysis/ast-evidence/journal--ReadOnly.BrokerOrderExitLinks.base.ast.json`(편집 전, `tools/logic-map`, 28 갈래 · 12 return,
source_sha256 26032b49…). 줄 번호는 편집 전 `internal/journal/account_views.go`.

- F1 행 수: SELECT 목록 `:450` `(SELECT count(*) FROM exit_events ce WHERE ce.proposed_intent_id=i.id) AS event_count` — action 구분 없음.
- F2 단일 행 연결: `:467-468` `LEFT JOIN exit_events e ON e.id=(SELECT min(ee.id) … WHERE ee.proposed_intent_id=i.id HAVING count(*)=1)` —
  행이 둘이면 서브쿼리가 0 행 → `e.*` 전부 NULL.
- F3 집계 갈래: AST B15 switch(`:534`)의 B21 case(`:545`) `eventCount > 1` → `unsafeReason = "ambiguous_exit_evidence"`;
  B22(`:548`) `!eventID.Valid` → `continue`; 루프 뒤 B26(`:572`) `agg.unsafeReason != ""` → `Ambiguous = true`.
- F4 훅 행의 모양: `appendExitEventFromHook`(apply_hook.go, 체결 훅 `PROPOSAL_FILLED`)와 `clearExitProposalTx`(apply_hook.go, 해제 훅
  `PROPOSAL_REFUSED` · `PROPOSAL_CANCELLED`)는 증거 열 없이 같은 `proposed_intent_id` 로 행을 씀. 무장 행은 `appendExitEventTx`(exit_state.go).
- F5 훅 목록 정의: `a207HookActions`(`internal/journal/a207_compact_evidence.go:16`) — OPENED · READOPTED · PROPOSAL_REFUSED · PROPOSAL_CANCELLED ·
  PROPOSAL_FILLED · COMPLETED · ADJUSTMENT_CLOSED 를 SQL 리터럴 목록으로 둔 패키지 상수. a207 의 「직전 무장」 조회가 같은 방식(`coalesce(e.action,'') NOT IN (…)`)으로 씀.
  `ADJUSTMENT_CLOSE_FILL_RECOGNISED`(a200)는 목록에 없지만 `proposed_intent_id` 를 쓰지 않음(a200_adjustment_close_fill_recognition.go:147) — 영향 없음.
- F6 호출자: `internal/console/orders.go` `Console.ledgerOrderEvidence`(:716) 하나. 결과는 `attachOrderExitEvidence` 가 화면 문구로만 바꿈.
- F7 운영 저널(읽기 전용 `mode=ro`, 2026-11-28, 행 수만): intent 당 (무장, 훅) = (1,1) 26,171 · (1,0) 5 · 그 밖 0. 훅은 REFUSED 13,041 ·
  CANCELLED 12,996 · FILLED 134. 압축 무장 행(`recomputed_repeat`)과 FILLED 가 같은 intent 인 경우 0.

## 2. 결정

- D1 F1 · F2 의 서브쿼리 둘에 `AND coalesce(<alias>.action,'') NOT IN (` + `a207HookActions` + `)` 을 더함. 목록 사본 0.
  `coalesce` 는 a207 조회와 같은 관례이며, action 이 NULL 인 행은 무장 행으로 셈 → 정체 불명 행이 하나 더 있으면 모호(fail-closed 쪽).
- D2 집계 갈래(F3)는 그대로. 무장 행이 둘 이상이면 F1 이 2 이상 → B21 → 모호. 시험 `TestA209TwoArmingRowsStayAmbiguousWithAHookRow` ·
  기존 `TestReadOnlyMarksDuplicateExitEvidenceAmbiguous`(LEGACY_A · LEGACY_B 는 훅이 아니므로 계속 둘로 셈)가 못 박음.
- D3 바뀌는 표시: (1,1) 은 모호 → 무장 행의 근거. (0,n)(훅 행만) 은 「legacy_event 로 연결」→ `exit_evidence_unlinked`(연결 없음) — 훅 행은
  근거가 아니므로 이쪽이 정직함. 운영 저널에 (0,n) 은 0 건(F7)이라 화면 변화 없음.
- D4 기존 함수 본문 편집은 `ReadOnly.BrokerOrderExitLinks` 하나(SQL 문자열 두 줄). 시험은 새 파일 `a209_exit_links_hook_rows_test.go` 의 새 함수로만
  더하고 기존 시험 함수는 고치지 않음 → gate ⑤ required 1(그 함수) — FLM 번들 `analysis/function-logic/internal-journal--readonly.brokerorderexitlinks/`.

## 3. 시험

| 시나리오 | 시험 | 모양을 만드는 방법 |
|---|---|---|
| 무장 + PROPOSAL_FILLED → 근거 표시 | `TestA209AFilledExitOrderLinksItsArmingRowDespiteTheFillHook` | `RecordExitJudgement` → `place`(매도 접수 확정) → `RecordFill`(종결) — 체결 훅이 실제로 씀 |
| 무장 + REFUSED / CANCELLED → 모호 아님 | `TestA209AReleasedExitOrderLinksItsArmingRowDespiteTheReleaseHook` | 같은 무장 · 주문 → `ResolveExitProposal`(생산 해제 쓰기 `clearExitProposalTx`) |
| 무장 둘 + 훅 → 모호 | `TestA209TwoArmingRowsStayAmbiguousWithAHookRow` | 위 체결 뒤 둘째 무장 행을 직접 심음(원장이 만들지 않는 모양) |
| 무장 둘(훅 없음) → 모호 | 기존 `TestReadOnlyMarksDuplicateExitEvidenceAmbiguous` | 무변경 |
| 압축 무장 하나 → 반복 사유 | 기존 `TestA207ReadersReadACompactRowAsARepeat` | 무변경 |
| 훅 행만 → 연결 없음(D3) · NULL action 행 + 무장 → 모호(D1) | `TestA209HookOnlyIsUnlinkedAndANullActionRowStillCounts` | 행 직접 심음(리뷰 r1 testing 지적으로 추가). `coalesce` 제거 변이가 둘째 갈래로 실패함(review.md) |

각 시험은 픽스처가 실제 두 행 모양인지(`a209ActionsOf`)를 먼저 단언함 — 기존 픽스처가 무장 행만 심어 결함을 놓친 것을 되풀이하지 않기 위함.

## 4. 같은 `proposed_intent_id` 로 묶는 다른 리더(읽기 전용 점검 — 고치지 않음)

| 리더 | 모양 | 훅 행을 세는가 | 판정 |
|---|---|---|---|
| `engineSellOrders`(trade_outcomes.go:615) | `exit_events` ⋈ `mutation_attempts` 로 `SELECT DISTINCT broker_order_id…` | 훅 행도 조인되지만 같은 intent → 같은 주문이라 DISTINCT 가 흡수. 세지 않음 | 결함 없음 |
| `HasProtectiveExitAttempt`(a087_protective_exit_attempt.go:26) | `EXISTS … e.action IN (<보호 action>)` | action 으로 거르므로 훅 행은 안 걸림(호출자가 훅 action 을 넘기지 않는 한) | 결함 없음 |
| provenance `exit_intent`(provenance.go:239) | `SELECT DISTINCT e.proposed_intent_id` | DISTINCT — 세지 않음. EXIT_EVENT 갈래는 이력 나열이라 훅 행이 보이는 것이 맞음 | 결함 없음 |
| `ExitProposalForIntent`(a166_replay_origin.go) | Go `switch` 로 훅 7 종 `continue` | 이미 건너뜀(목록은 Go 상수 나열 — SQL 상수와 별개 사본) | 결함 없음 |

저장소 전체 `rg 'count\(\*\) FROM exit_events|HAVING count'` 에서 `exit_events` 를 intent 별로 세는 자리는 이 리더의 두 곳뿐이었음.

## 5. 열린 질문(사람 결정)

1. **압축 무장 행의 체결 주문 표시.** a207 압축 행(`effective_source='recomputed_repeat'`)이 무장한 intent 가 체결되면, 이제 모호가 아니라
   `repeat_of_refused_proposal`(스냅숏 없음)으로 보임. D8 앵커(같은 포지션의 직전 전체 증거 무장 행)를 따라가 근거를 보여 줄 수도 있음.
   **기본(이 change): 바꾸지 않음** — 앵커 조회를 리더에 더하는 것은 새 기능이고, 운영 저널에 그 모양이 0 건(F7)임.
   단 리뷰(적대 보이스 #3)가 짚은 대로, a207 배포 뒤에는 막힘이 풀린 뒤 체결되는 무장이 압축 행인 경우가 흔해질 수 있음(같은 키의
   거절 다음 무장은 압축됨) — 그때 체결된 주문에 `operatorview` 의 「반복 거절(요약) — …」 문구가 뜸. 매매 영향 0, 문구 오독 위험만 있음.
2. **a207 spec 델타와의 충돌.** 아직 archive 되지 않은 a207 의 exit-policy 델타 Requirement 「리더는 압축 증거 행을 오류 없이 읽는다」는
   「orders 연결의 연결 행 수 판정은 바뀌어서는 안 된다(SHALL NOT)」와 Scenario 「연결 행이 둘 이상이다 — 한 intent 에 압축 행과 체결 훅 행이 함께
   연결되면 지금처럼 모호」를 담고 있음. 이 change 는 바로 그 행 수 판정을 고치므로, a207 이 그대로 archive 되면 main spec 에 정반대 문장이 남음.
   a207 은 다른 writer 의 change 라 여기서 고치지 않음. 선택지: (가) a207 archive 전에 a207 소유자가 그 문장 · Scenario 를 「무장 행 수」로 고침,
   (나) a207 archive 뒤 a209 가 exit-policy MODIFIED 델타를 더함. 권장 (가) — 문장 둘, archive 순서와 무관.

## 6. 롤백

커밋 되돌리기 + `make image` 재빌드. 스키마 · 데이터 변경 0, 설정 0.
