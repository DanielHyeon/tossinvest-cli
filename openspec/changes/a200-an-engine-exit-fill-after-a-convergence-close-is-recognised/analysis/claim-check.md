# a200 1.6 · 1.11 — design 전제와 주장 대조

base `884be2a74`. AST 좌표는 `analysis/cited-ast/*.ast.json`(`B<n>(kind@줄)`), 재현은 `analysis/repro/`. [손] = SQL · DDL · 리터럴을 손으로 읽음.
판정: **참** / **거짓** / **미확인**(결정이 필요하거나 이 단계에서 잴 수 없음).

## 1.6 design 전제 (a)~(f)

| # | 전제 | 판정 | 근거 |
|---|---|---|---|
| (a) | `SELL_ON_CLOSED` 는 `resolveFillOrigin` 통과 뒤에만 생긴다 | **참** | `ProjectPosition` resolveFillOrigin@66 → B2(if@70) `!found` → return(@75). 거절 갈래 B5(if@101) → enterReconcileInTx@102 는 그 뒤에만. 재현 W4 의 매도는 접수 확정 주문 |
| (b) | `exit_events.action` 에 스키마 제약이 없다 | **참(action 한정)** | DDL `core_domain.go:251-261` [손]: `action TEXT` — CHECK 없음, FK 는 `position_id REFERENCES positions(id)` 하나. 뒤 판본은 열 추가(`position_policy.go:23`, `exit_snapshot.go:40-59`)와 인덱스 셋(`idx_exit_events_position`, 부분 `idx_exit_events_proposed_intent`, UNIQUE 부분 `idx_exit_events_decision` ON decision_id) 뿐, `exit_events` 트리거 0(grep). **단 D3 의 「인정 수량을 실음」 은 실을 열이 없음** — 아래 C-D3 |
| (c) | Project 훅 시점에 `pending_intent_id` 가 살아 있다(부분 · 종결) | **참** | 훅 순서 `runApplyHooks` Project@284 → Exit@298. Exit 훅은 B11(if@1009: Terminal ∧ 일치)에서만 ResolvePending@1011. 조정 tx 는 pending 을 안 건드림. 재현: 조정 직후 `i-stop`, Working 체결 뒤 `i-stop`, Done 체결 뒤 빈 값(`w4.txt`) |
| (d) | `pending_intent_id` 쓰기 경로 전수 | **사실 확정** | 값을 쓰는 곳 2 함수 · 생산 호출 2 자리, 둘 다 exit 루프: `armExitProposalTx`(apply_hook.go:681 ← RecordExitJudgement(Result) ← exitloop.go:1426, IntentID 는 :1421 에서 무장 시 채움), `AttachExitIntent`(:800 ← exitloop.go:1540). 비우는 곳 3 함수: `ApplyTx.ResolvePending`(← ApplyExitFill:1011), `clearExitProposalTx`(← `ResolveExitProposal` 생산 호출 0 · `ReleaseUnacceptedExitProposal` 생산 호출 4 — attempt_thaw_command.go:121, boot_catchup.go:62, exitloop.go:1624, exit_held_proposal.go:141, 모두 「접수 안 된」 시도만), `resetExitStateForReadoptTx`(← position_policy.go:145). **flatten · 운영자 매도 · 편입은 0 자리** — 그 매도는 구조적으로 (다)를 못 통과(design 비목표와 일치) |
| (e) | 조정 `kind` 값 전수와 수렴이 쓰는 값 | **사실 확정** | CHECK `kind IN ('EXTERNAL','MANUAL','UNKNOWN')`(core_domain.go:161 [손]). 생산 쓰기 2 자리: 수렴 converge.go:213-215 `Classify(LocalInstance:true)` = UNKNOWN, 외부 접기 external.go:226 EXTERNAL(기대 이전 0 → 새 인스턴스를 엶 — 재현 seq 2). MANUAL 생산 쓰기 0(`OperatorDeclared` 를 참으로 주는 생산 코드 없음). ⇒ 기존 인스턴스를 0 으로 닫는 조정은 지금 수렴(UNKNOWN)뿐. §2 「계속 거절 6」 은 생산에서 만들 입력이 없고 원장 직접 쓰기로만 시험 가능 |
| (f) | 체결이 「주문이 결속된 인스턴스」 를 찾는가 | **거짓** | 투영은 `currentInstance`(@78, SQL 271-276 [손] `WHERE account_ref, market, symbol ORDER BY instance_seq DESC LIMIT 1`), Exit 훅도 `exitContextForFill`(939-942 [손] 같은 SQL) — **둘 다 최신 인스턴스**. 주문 → 인스턴스 결속 열 없음. 재현: 재매수가 새 EXTERNAL 인스턴스로 접힌 뒤 옛 손절 31 체결 → 새 인스턴스(5)에 닿아 X50 OVERSELL → QM; 부분 3 체결 → **차단 없이 새 보유 5 → 2**(`w4.txt`). ⇒ 그 체결은 `SELL_ON_CLOSED` 에 오지 않으므로 인정 판정에 닿지 않음 |

## 1.11 주장 대조

| id | 주장(문서) | 판정 | 근거 · 설명 |
|---|---|---|---|
| C-W | proposal Why 1~4, 운영 집계의 W4/W5/W6 서술 | **참** | 재현 W4 · W5 · W6(`cycles.txt`) — 서술과 같은 결과 |
| C-D1 | 체결 tx 와 조정 tx 는 직렬이고 CLOSED 를 읽었다는 것이 「조정이 먼저 커밋」 의 증명 | **참** | `db.SetMaxOpenConns(1)`(journal.go:174) + `_txlock=immediate`(:225) [손]. 그리고 CLOSED 는 종결 상태라(아래 C-D2나) 순서를 시각으로 비교할 필요가 없음. 재현에서 조정과 체결이 같은 초 |
| C-D2가 | 거절 유형 `SELL_ON_CLOSED` 로 가른다 | **참** | 판정 자리에서 `outcome.Refusal`(position.Refusal 값)을 그대로 씀 — evidence 문구 해석 불요(a195 R2 의 「유형은 문구에만」 은 원장 행 얘기) |
| C-D2나 | 「그 인스턴스의 가장 최근 수량 변화가 수렴 조정이고 0 으로 만들었다」 를 원장에서 판정할 수 있다 | **참(방법 단서)** | 조정 행과 체결 행의 순서를 id 로 비교할 수 없고(다른 표 · `fill_events` 에 position_id 없음) 시각도 초 단위지만, **필요 없음**: CLOSED 는 종결 — 전이표 table.go:59-61 · 106-108(이동 없음 → CLOSED 유지), 162-164(초과 매도 → 거절), 83-85(매수 → NewInstance) [손]. 수렴은 CLOSED 를 조정하지 않고(`soleLiveMarket` 「no live instance」, 재현 W2), EXTERNAL 접기는 새 인스턴스를 엶(재현). ⇒ 「그 인스턴스의 조정 중 new_quantity=0 인 행이 있다」 ⟺ 「조정이 닫았다」. 체결로 닫힌 뒤 둘째 매도(재현)는 조정 행 0 이라 걸러짐 |
| C-D2다 | 체결 intent = 그 인스턴스 `pending_intent_id` | **참(구현 제약)** | 값은 살아 있음((c)). 단 `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`(apply_hook_test.go:567)이 apply_hook.go · core_domain.go · position_policy.go 밖 생산 파일이 `pending_intent_id` 철자를 **언급만 해도** 실패시킴 → 새 판정 파일은 `ApplyTx.PendingState`(apply_hook.go:367)로 읽어야 함 |
| C-D2라 | 인정 누적을 `exit_events` 로 세는 것이 부분 · 재시도 · 재적용에 안전 | **참(누적) / 미확인(저장)** | 부분: 체결마다 delta = 누적 차, 거절돼도 스냅숏 전진(재현 delta 20 → 11). 재적용: 바이트 같은 재관측은 훅 전에 반환(fills.go 「byte-identical re-observation touches nothing」), delta 0 비동일 관측은 훅을 돌지만 CLOSED 이동 없음(X16-18)이라 거절 · 인정 없음. 훅 오류는 tx 전체 롤백 → 기록과 스냅숏이 함께. **수량을 담을 열이 없음** — C-D3 |
| C-D3 | 인정 기록 = `exit_events` 새 action 한 줄, 인정 수량을 실음, 스키마 36 | **거짓(수량 열)** | 수량 모양 열은 `projected_quantity` 뿐인데 평가 증거 묶음의 일부라, 그것만 채우면 `exitEventEvidence.any()` 참 · `full()` 거짓 → `ExitEvents` 읽기가 `ErrExitSnapshotCorrupt`(exit_snapshot.go:376-401) — 그 포지션 이력 읽기 전부가 깨짐. 남는 열(`observed_price` · `high_water` · `baseline_after` · `level_after`)은 콘솔 · httpapi 가 가격 · 레벨로 보여 줌(account_views.go:311 · 452). 또 새 행에 `proposed_intent_id` 를 실으면 `ExitProposalForIntent`(journal/a166_replay_origin.go:38 의 수명 이벤트 목록 밖 → `invalid`)가 그 손절 intent 의 출처를 「없음」 으로 바꿔 a166 예산 우선순위가 Critical → Normal(execgw/a166_replay_origin.go:47-52). → **Manager 결정** |
| C-D4 | 부분 체결은 체결마다 판정, Project 시점에 pending 살아 있음 | **참** + 단서 | (c) 와 같음. 단서: 「주문 종결」 신호가 delta 0 관측으로 늦게 올 수 있고 그 관측은 거절 갈래를 안 지남 → D5 완결 판정은 ApplyExitFill 에서 원장의 인정 누적 + `fill.Terminal` 로 해야 함(거절 순간에만 하면 놓침) |
| C-D5a | 엔진 진입 포지션: 성과 입력은 그 주문의 체결 기록으로 충분 | **참(같은 초)** | W4 원장 동결 탐침 = W0 동결 값(PnL · R · 비용 · 단계, `w4.txt`). `roundTripLegs` 가 거절된 매도 `fill_events` 도 합산. 단 `closed_at` · `held_seconds` 는 `positions.closed_at`(= 조정 시각, computeTradeOutcome SQL 184 · heldSeconds 261 [손])이라 조정과 체결이 다른 초면 W0 와 다름 → design §3-6 「W0 와 같다」 는 같은 초에서만 참 |
| C-D5b | 편입 포지션도 성과 동결(§2 정상 5 · §3) | **거짓** | 탐침 `post_adoption_disposal`(`w4.txt`): `adoptionEpoch` B8(switch@229) → B11(case@238) → B12(if@240) 편입 뒤 감소 → B14(if@250) 거절. 수렴 조정 자체가 「편입 뒤 감소」. 고치려면 `adoptionEpoch`(기존 함수 3번째 — 백필 계획 · 묶음 목록도 부름)를 편집해야 함 → **멈춤 조건** |
| C-D5c | 정본과의 정합 | **거짓(MODIFIED 필요)** | position-ledger 「Position 투영과 단일 권위」 본문 「조정 이벤트로 포지션 수량이 0이 되면 … trade_outcome 행은 만들지 않는다(SHALL NOT)」 + Scenario 「외부 매도로 수량 0」(편입 포지션 · 성과 행 없음). trade-analytics 「성과 원시 지표」 의 「체결로 거래 완결」. trade_outcome_backfill.go:12-16 주석의 「COMPLETED 와 ADJUSTMENT_CLOSED 는 한 인스턴스에서 배타적」(백필 모집단 근거)도 D5 뒤 거짓 |
| C-D6 | 인정돼도 수량 불일치 차단은 크레딧으로 풀림 | **참(등가 재현)** | 인정 뒤 Observe 시점 원장은 「AF 행 없음 + 조정 · 크레딧 있음」 = W6 의 상태. W6 재현에서 QM 이 첫 일치에 `ADJUSTMENT_APPLIED`(`cycles.txt`). 수정 후 코드로 W4 · W5 를 돈 것은 아님(구현 단계 §3-5) |
| C-D7 | a130 `OutcomeRefused` 식 통로로 로그 | **맞지 않음** | 그 통로의 값은 `TradeOutcomeRefusal{Reason}` 이고 사유 집합은 닫혀 있음(`TestA130TheReasonSetIsClosedAndTokenShaped`). 새 콜백이면 `ApplyHooks` 필드 + `bindApplyHooks` 본문(a130 시험이 인자 수 2 를 단언). design 의 대안대로 **로그 생략 · 원장 기록만** 권고 |
| C-§2-5 · 정상3 | 재매수 뒤 옛 주문 체결은 「주문이 결속된 인스턴스」 로만 판정 / 정상 입력으로 통과 | **거짓** | (f). 그 체결은 거절 갈래에 오지 않음. 지키려면 `currentInstance` · `exitContextForFill` 의 인스턴스 해석을 바꿔야 함(기존 함수 2 개 더 + 인스턴스 귀속 규칙 변경) → **Manager 결정** |
| C-§2-정상4 | `MissingOrders` 로도 잡힌 경우 | **참** | W4 · W5 재현의 비교에 `MissingOrders` 포함 — 체결 쪽 판정과 무관 |
| C-§2-정상6 | W6 에서도 같은 판정 | **참** | W6 에서는 활성 QM 때문에 AF 를 안 씀(F10) — 인정 판정은 행 쓰기 전 자리라 같은 판정을 하되 결과(차단 없음)는 오늘과 같음 |
| C-§5 | 편집할 기존 함수 둘(`ProjectPosition` · `ApplyExitFill`) | **거짓(조건부)** | 둘로 되는 조건: D5 를 엔진 진입에 한정 · 인정 행에 `proposed_intent_id` 를 안 실음 · 수량 열 문제를 기존 함수 편집 없이 풂 · (f) 를 범위 밖으로. 아니면 `adoptionEpoch` · `ExitProposalForIntent` · `currentInstance` · `exitContextForFill` 중 필요한 만큼 늘어남 → **멈춤 조건** |
| C-비목표 | 「W2 · W3 는 코드 추론이고 재현 전」 | **정정함** | 재현됨(`out-of-scope.md`) — proposal 비목표 · design §4 의 문장만 사실대로 고침(범위 결정은 그대로) |
| C-빠진 소비자 | design 이 다루지 않은 소비자 | **발견** | a137 일일 손실 측정: 수렴으로 닫힌 관리 포지션에 엔진 매도가 발송됐으면 `ErrUnmeasurable`(dailyloss/measure.go 124-128) — 그 갈래가 성과 행보다 먼저라 **a200 뒤에도 그날 측정 불가**(Guardian 일일 손실 검사가 켜져 있으면 그날 진입 거절, execgw/a137_daily_loss.go). `edit-surface.md` 소비자 표 |
