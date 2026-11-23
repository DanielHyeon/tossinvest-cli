# a200 1.6 · 1.11 — design 전제와 주장 대조 (범위 개정 2026-11-24 판)

base `884be2a74`. AST 좌표는 `analysis/cited-ast/*.ast.json`(`B<n>(kind@줄)`), 재현은 `analysis/repro/`. [손] = SQL · DDL · 리터럴을 손으로 읽음.
판정: **참** / **거짓** / **미확인**. 첫 판(개정 전 design)의 거짓 다섯은 아래 「이력」 에 남기고, 개정 design 의 주장은 「1.11」 표에서 다시 대조한다.

## 1.6 design 전제 (a)~(f)

| # | 전제 | 판정 | 근거 |
|---|---|---|---|
| (a) | `SELL_ON_CLOSED` 는 `resolveFillOrigin` 통과 뒤에만 생긴다 | **참** | `ProjectPosition` resolveFillOrigin@66 → B2(if@70) `!found` → return(@75). 거절 갈래 B5(if@101) → enterReconcileInTx@102 는 그 뒤에만. 생산에선 RecordFill 이 소유 확정일 때만 훅을 부름(fills.go:550). 재현 W4 의 매도는 접수 확정 주문 |
| (b) | `exit_events.action` 에 스키마 제약이 없다 | **참** | DDL `core_domain.go:251-261` [손]: `action TEXT` — CHECK 없음, FK 는 `position_id REFERENCES positions(id)` 하나. 뒤 판본은 열 추가(`position_policy.go:23`, `exit_snapshot.go:40-59`)와 인덱스 셋(`idx_exit_events_position`, 부분 `idx_exit_events_proposed_intent`, UNIQUE 부분 `idx_exit_events_decision` ON decision_id) 뿐, 트리거 0(grep). 개정 design 은 수량을 싣지 않음 |
| (c) | Project 훅 시점에 `pending_intent_id` 가 살아 있다(부분 · 종결) | **참** | 훅 순서 `runApplyHooks` Project@284 → Exit@298. Exit 훅은 B11(if@1009: Terminal ∧ 일치)에서만 ResolvePending@1011. 조정 tx 는 pending 을 안 건드림. 재현: 조정 직후 `i-stop`, Working 체결 뒤 `i-stop`, Done 체결 뒤 빈 값(`w4.txt`) |
| (d) | `pending_intent_id` 쓰기 경로 전수 | **사실 확정** | 값을 쓰는 곳 2 함수 · 생산 호출 2 자리, 둘 다 exit 루프: `armExitProposalTx`(apply_hook.go:681 ← RecordExitJudgement(Result) ← exitloop.go:1426, IntentID 는 :1421 에서 무장 시 채움), `AttachExitIntent`(:800 ← exitloop.go:1540). 비우는 곳 3 함수: `ApplyTx.ResolvePending`(← ApplyExitFill:1011), `clearExitProposalTx`(← `ResolveExitProposal` 생산 호출 0 · `ReleaseUnacceptedExitProposal` 생산 호출 4, 모두 「접수 안 된」 시도만), `resetExitStateForReadoptTx`(← position_policy.go:145). **flatten · 운영자 매도 · 편입은 0 자리** — 그 매도는 구조적으로 (다)를 못 통과 |
| (e) | 조정 `kind` 값 전수와 수렴이 쓰는 값 | **사실 확정** | CHECK `kind IN ('EXTERNAL','MANUAL','UNKNOWN')`(core_domain.go:161 [손]). 생산 쓰기 2 자리: 수렴 converge.go:213-215 `Classify(LocalInstance:true)` = UNKNOWN, 외부 접기 external.go:226 EXTERNAL(기대 이전 0 → 새 인스턴스 — 재현 seq 2). MANUAL 생산 쓰기 0. ⇒ 기존 인스턴스를 0 으로 닫는 조정은 지금 수렴(UNKNOWN)뿐. §2 「계속 거절 6」 은 지어낸 MANUAL 행으로만 시험 |
| (f) | 체결이 「주문이 결속된 인스턴스」 를 찾는가 | **아니다(사실)** — 개정 design 은 이를 범위 밖 사실로 적음 | 투영은 `currentInstance`(@78, SQL 271-276 [손]), Exit 훅도 `exitContextForFill`(939-942 [손]) — 둘 다 (계좌 · 시장 · 종목)의 최신 인스턴스. 재현: 재매수 뒤 옛 손절 전량 → 새 인스턴스에서 X50 OVERSELL → QM, 부분 → 차단 없이 새 보유 5 → 2. 그 체결은 인정 판정에 닿지 않음(design §2 「범위 밖 사실」) |

## 1.11 개정 design 주장 대조

| id | 주장(개정 문서) | 판정 | 근거 |
|---|---|---|---|
| C-W | proposal Why 1~3 · W4/W5/W6 서술 | **참** | 재현 W4 · W5 · W6(`cycles.txt`) |
| C-D1 | 체결 tx 와 조정 tx 직렬, CLOSED 를 읽음 = 조정 먼저 | **참** | `db.SetMaxOpenConns(1)`(journal.go:174) + `_txlock=immediate`(:225) [손]. 재현에서 같은 초 |
| C-D2가 | `outcome.Refusal == SELL_ON_CLOSED` 로 가름 | **참** | 판정 자리에서 `position.Outcome.Refusal` 값을 그대로 씀 — evidence 문구 해석 불요 |
| C-D2나 | 「0 조정 행이 정확히 하나 · 마지막 조정 행 · kind UNKNOWN」 ⟺ 수렴이 닫음, 조정 · 체결 순서 비교 불요 | **참** | CLOSED 는 종결: 전이표 table.go:59-61 · 106-108(이동 없음 → CLOSED), 162-164(초과 매도 → 거절), 83-85(매수 → NewInstance) [손]. 수렴은 CLOSED 를 조정 안 함(`soleLiveMarket`, 재현 W2), EXTERNAL 접기는 새 인스턴스(재현). 같은 조정의 재적용은 파생 id 로 행을 안 만듦(adjustment_close_test `…DoesNotCloseTwice`). 체결로 닫힌 뒤 둘째 매도(재현)는 0 조정 행이 없어 걸러짐 |
| C-D2다 | `origin.IntentID` = `PendingState(인스턴스).PendingIntentID` | **참** | 값은 살아 있음((c)). `PendingState`(apply_hook.go:367)는 exit 상태가 없으면 `ErrExitStateNotFound` → 인정 아님(무관리 인스턴스). 새 파일이 `pending_intent_id` 철자를 쓰면 `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` 실패 — 그래서 `PendingState` |
| C-D2a | 누적 합 출처 = 주문별 **브로커 원문 누적**(이 주문 `fill.CumulativeQuantity`, 다른 주문 `fill_snapshots.filled_quantity` 최댓값)의 decimal 합, 대상 = 그 intent 의 CONFIRMED PLACE/AMEND 주문 번호들, 범위 필터 없음 | **참(r1 뒤 개정)** | 첫 개정판의 `fill_events.delta_quantity` 합은 **거짓이었음**(codex r1 P1-1): delta 는 float64 차이고 `nearlyZero`(fills.go:1924, 상대 1e-9)가 작은 차를 0 으로 만들어 행을 안 남기는데 스냅숏은 전진 → 합이 실제 누적보다 작아질 수 있음. 원문 누적은 그 손실이 없음(스냅숏 upsert 가 `filled_quantity = excluded.filled_quantity`, fills.go:742 등 [손]). 바이트 같은 재관측은 훅 전에 반환(fills.go:417-425), delta 0 관측은 `classifyMovement` 0 → MoveNone(position.go:293-295) → 판정 안 부름. 범위 필터를 빼고 최댓값을 쓰면 오차는 더 세는 쪽뿐(인정이 줄어드는 방향). 정정 자식이 다른 intent 면 (다)에서 거절 |
| C-D2a-범위 | 「다시 유도하지 않음」 | **참** | 소유 intent 는 `origin.IntentID` 를 받음. 주문 번호 목록은 그 intent 의 attempt 를 읽을 뿐 소유 판정(범위 일치 · 유일 소유 · `settled_at < committed_at`)을 다시 하지 않음 — 그 판정을 빼서 생기는 차이는 더 세는 쪽 |
| C-D3 | 표지 행이 모든 소비자에게 무시됨 | **참(원문 표시 둘 포함)** | `edit-surface.md` 소비자 표 — 판정 소비자 전부 무시(intent · decision · action 키), 증거 검증은 legacy 갈래, 화면 둘(콘솔 이력 · provenance)은 원문 한 줄 표시. 오분류 · 깨짐 0 |
| C-D4 | 부분 체결 체결마다, Project 시점 pending 살아 있음 | **참** | (c) |
| C-D5 | Exit 훅 · 성과 무변, 인정된 청산은 `ADJUSTMENT_CLOSED` + 표지 + `PROPOSAL_FILLED` | **참** | `ApplyExitFill` 무편집. 오늘 W4 원장의 exit_events 가 `ADJUSTMENT_CLOSED · PROPOSAL_FILLED`(재현) — 표지는 Project 훅에서 그 사이에 들어감 |
| C-D6 | 인정돼도 수량 불일치 차단은 크레딧으로 다음 일치에서 풀림 | **참(등가 재현)** | 인정 뒤 Observe 시점 원장 = W6 의 상태(AF 행 없음 + 조정 · 크레딧). W6 재현에서 첫 일치에 `ADJUSTMENT_APPLIED`, 진입 열림. 단 크레딧은 프로세스 기억 — 재기동이 끼면 「조정 대기」 로 남음(codex r1 P2-2, `Tracker.Restore` mismatch.go:998) → design D6 · delta 에 조건으로 적음. 수정 후 코드로 W4 · W5 를 돈 것은 구현 단계(design §3-4) |
| C-§2 대가 | 조정을 사이에 둔 부분 체결은 **누적이 감소분을 넘는 체결부터** AF(결국 AF) | **참(r1 뒤 정정)** | 첫 개정판 「모두 거절」 은 거짓이었음(codex r1 P2-1: 조정 뒤 1 · 20 으로 나누면 첫 체결은 누적 11 ≤ 21 로 인정). 재현 `…StraddlingTheConvergence…`(나머지 21 한 번) → X59 AF, 합 31 > 21. 주문이 끝까지 체결되면 누적 = 원주문 > 감소분이라 마지막 체결은 반드시 AF |
| C-인과 | D2 를 통과한 체결은 그 조정의 감소를 일으킨 매도다 | **미확인 — Manager 결정 필요** | codex r1 P1-2 반례: 엔진 매도 대기 중 외부 처분으로 조정 31→0, 재매수 31 이 원장에 접히기 **전에** 옛 엔진 매도 31 체결 → D2 전부 통과하지만 그 매도는 재매수분의 처분. 원장에는 조정 스냅숏과 체결의 인과를 가를 열이 없음. 실현 가능성은 브로커가 미체결 매도에 묶인 수량의 외부 매도를 허용하는지에 달림(미측정) |
| C-W6 예외 | 「어긋나면 AF 를 만든다」 | **참(r1 뒤 정정)** | 활성 행이 있으면 안 씀(F10, position_projection.go:401) — reconciliation delta 를 그 조건으로 좁힘(codex r1 P2-3) |
| C-§2 정상 1~4 | 같은 초 · 조정 뒤 부분 여러 번 · `MissingOrders` 동반 · W6 | **참** | 같은 초(재현), 부분 여러 번은 원문 누적이 20 → 31 로 감소분 31 이하(재현), `MissingOrders` 는 체결 쪽 판정과 무관(재현 비교에 포함), W6 은 판정이 같고 차단은 원래 F10 으로 안 씀 |
| C-§5 | 편집할 기존 함수 하나(`ProjectPosition`) | **참** | 판정 · 표지는 새 파일. 쓰는 기존 함수(`PendingState` · `appendExitEventFromHook` · `ApplyTx.Query`)는 호출만. pre-edit 번들 `function-logic/internal-journal--projectposition/` |
| C-정본 | 정본과 모순 없음 | **참(delta 뒤)** | position-ledger 「포지션 상태기계」 의 「허용되지 않은 전이는 … RECONCILE」 이 유일한 모순 → MODIFIED(예외 한 문장, `modified-delta/build_delta.py` 역치환 단언). 성과 문장(「외부 매도로 수량 0 … trade_outcome 행은 생성되지 않는다」 · trade-analytics)은 인정된 청산도 성과 0 이라 모순 없음. reconciliation 「귀속 실패는 스스로 풀리지 않는다」 는 인정이 차단을 **안 만들 뿐** 풀지 않으므로 모순 없음. risk-management:654(투영 거절 종목 세기) · multi-horizon-risk-buckets:118(late sell owner)은 표현 대상이 다름 — 무변 |
| C-비목표 | W2 · W3 재현됨, a137 측정 불가 그대로, 운영 로그 없음 | **참** | `out-of-scope.md` · dailyloss/measure.go:124-128 · a130 사유 집합 닫힘(`TestA130TheReasonSetIsClosedAndTokenShaped`) |

**잔여 거짓: 0. 미확인 1(C-인과 — codex r1 P1-2, Manager 결정).**

## 이력 — 첫 판(개정 전) 대조에서 거짓이던 것과 처분

| 첫 판 주장 | 판정 | 처분(Manager 2026-11-24) |
|---|---|---|
| D3 「인정 수량을 `exit_events` 에 실음」 | 거짓 — 실을 열 없음(`projected_quantity` 는 평가 증거 묶음, 단독으로 채우면 `ExitEvents` corrupt) | 수량 저장 안 함, (라)를 상태 없는 합으로(결정 2) |
| D5 편입 포지션 성과 동결 | 거짓 — `post_adoption_disposal`(재현 탐침) | 성과 비목표(결정 1) |
| D5 · 정본 정합 | 거짓 — 「외부 매도로 수량 0 … 성과 행 없음」 과 충돌 | 성과 비목표라 해소 |
| §2 계속 거절 5 · 정상 입력 3 | 거짓 — 전제 (f) | 범위 밖 사실로(결정 5) |
| §5 편집 함수 둘 | 거짓(조건부) | 하나로(결정 9) |
| D7 a130 통로 로그 | 맞지 않음 | 뺌(결정 4) |
