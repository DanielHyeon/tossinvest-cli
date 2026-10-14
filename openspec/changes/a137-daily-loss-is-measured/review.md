# a137 review

## 1. D0 증거 독립 대조 (2026-10-14, 분리된 읽기 전용 컨텍스트 — 판정은 Manager 가 전달)

판정: **ACCURATE-WITH-CORRECTIONS** — P1×4 · P2×5 · P3. 전부 proposal · design 에 반영했다(design.md D0 의 「정정」 행).

| 발견 | 내용 | 반영 |
|---|---|---|
| P1-1 | F6/R9 하류는 레인 잠금이 아니다. 생산 감독자(`RefreshesAuthority: true`)는 `refreshOnly` 라 `recordSwallowedCycleError` + `continue`, `latchMarket` 미도달, 그 주기 뒤 범위 미평가, 폴링마다 재거절, 강화 없음 | design F7 · proposal 5 |
| P2-2 | admission.go:93 이 원인을 버리고 dispatch_cycle.go:157 은 Detail 만 포맷 → `errors.As` 불가 | design F8 · 첫 다리 FLM 편집 계획 |
| P1-3 | `IssueEntry` 의 생산 경로 0 → `escalateFor` 잠듦. 살아 있는 입구는 `PrecheckQFinalEntry` 하나 | design F6 |
| P1-4 | Q5(b)(측정마다 공식 환율 읽기)는 risk-management FX 요구(재조회 금지) 위반 | Q5 는 사용자가 같은 파도 권한 재사용으로 결정 → design D5 |
| P1-5 | a130 backfill 은 체결로 닫힌 편입 포지션만. 엔진 진입 포지션의 거절은 복구 경로 없음, 조정 종결 제외 | design F9 · 거부 목록 2 |
| P2-6 | 성과 PnL = 인스턴스 수명 전체(부분 익절 포함), `closed_at` 귀속 | design F9 · D4 U2 |
| P2-7 | 일일 손실 예약 관례는 생산에서 실행되지 않음. 살아 있는 선례 = 콘솔 `marketDayStart` · `AccountTradeTrips` | design F10 |
| P2-8 | U1 과 U3 의 조정 종결 처리 모순 | 사용자 결정 Q13 으로 해소 |
| P2-9 | 순실현이면 한도 뒤 이익이 같은 날 진입을 다시 연다 | 사용자 결정 Q12 로 해소 |
| P3 | B5/B7/B8 = INPUT_UNAVAILABLE · riskcalc B10 clamp · `closed_at` 은 trade_outcomes 칸 · counterfactual `"0"` · 체인 없는 `execgw.Issuer.IssueEntry` 생산 호출자 0 고정 · `AsOf ≤ Now` · precheck 1회 측정 · 단일 연결 · precheck 브로커 호출 금지 | design F1 · F3~F5 · F12 · D2 |

## 2. 사용자 결정 (2026-10-14, 최종 — Manager 전달 원문)

> Q1 = journal (fix riskcalc RealizedPnL.AsOf comment). Q2 = exclude partial closes of open positions (stated limit, follow-up).
> Q3 = fail-closed for that market day (with the P1-5 consequence stated: engine-entry refusals unblock only at the day boundary).
> Q4/day = KST 08:00 rollover ≡ each market's own trading date D, summed across markets — same definition as a136 D1; a136 will put
> the day-key function in one shared place; call it, don't duplicate. Q5/FX = reuse the USD/KRW authority already sealed by the US side
> in the same wave (no new broker call); check and state spec compliance; if it truly can't comply, stop and report.
> Q6 = classify strategy-path daily-loss refusal as account-level entry block and escalate to ENTRY_BLOCKED (preserve cause per P2-2).
> Q7 = keep (until human release). Q8 = exclude unrealized. NEW Q12 = count only LOSSES of positions closed that day, profits do not
> offset (conservative; measured total never decreases within a day). NEW Q13 = managed positions closed by the human selling in the app
> (adjustment-closed) are NOT counted as loss and do NOT fail closed (stated limit) — consistent with "manual orders' losses not counted".

Q5 FX 조항 승인(2026-10-14, 사용자 선택 「1」 — Manager 전달 원문):

> the user APPROVED (2026-10-14, choice "1") the narrowed FX MODIFIED clause for risk-management 「KR과 US는 하나의 account-base Guardian 권위를
> 공유한다」: base-currency losses need no conversion/authority; another market's loss/buy amount may be converted only with that market's authority
> carried in the same paired FX collection (same observedAt) and validated by Guardian at precheck (official source, pair, freshness); absent → fail
> closed for entries; re-read, caller-supplied rate, other-wave authority remain forbidden; request sizing/exposure/reservation/envelope FX rules unchanged.

순서 지시: a136 이 먼저 구현하고 공유 production 파일(`risk.Policy` · `guardian_wiring.go` · `riskguardian_first_leg.go` · a136 이 MODIFY 하는
risk-management 요구)을 소유한다. a137 은 그 커밋 전에는 새 파일만 만든다. base 는 a137 첫 Go 커밋의 부모.

## 3. Q5 스펙 준수 판정 — 승인됨

지금 문장 그대로는 준수하지 못했고(KR 요청의 aggregate 가 「다른 환율」을 씀), 좁은 MODIFIED 로 준수한다. 그 조항이 위 §2 원문대로 **승인되었다**(2026-10-14).
델타의 문구는 승인 원문에 맞췄다(base 통화 무환산 · 같은 paired FX collection · precheck 검증 · 없으면 fail-closed · 재조회/호출자 환율/다른 파도 금지 · 요청 FX 규칙 불변).
이제 이 조항에 기대는 코드를 써도 된다(구현은 여전히 a136 공유 파일 커밋 뒤).

## 4. proposal-freeze 리뷰 (2026-10-14, 분리된 읽기 전용 컨텍스트 4보이스 — 적대적 Eng · 스펙 · 범위 · 운영; 판정은 Manager 가 전달)

판정: **APPROVE-WITH-FIXES — P1 다섯이 닫힐 때까지 freeze 차단.** 아래는 전부 사용자 결정 안의 수리다(새 사람 결정 없음, P1-5 는 위 승인으로 해소).

| 발견 | 내용 | 처분 |
|---|---|---|
| P1-1 (fail-open) | 시장 현지일 귀속은 늦게 기록된 손실을 끝난 날로 떨어뜨린다(`closed_at` = 원장 커밋 시각, position_projection.go:316-318 · fills.go:174) | design D1-1 · F10: `clock.AccountTradingDay(closed_at) == D` 로 귀속. RED 두 행 뒤집음(US 03:00Z · KR 15:00Z → 들어감) + 경계 행 · 늦은 기록 시험 |
| P1-2 (fail-open) | Q13 제외는 사람의 앱 매도에만 — 수명 안에 발송된 엔진 SELL 이 있으면 측정 불가, `new_quantity` 는 decimal 비교 | design D1-4 · spec 시나리오 · RED `TestA137AnAdjustmentCloseWithADispatchedEngineSellIsUnmeasurable` |
| P1-3 | 서명 활성화된 레인은 `refreshOnly` 거짓 → (966) durable 레인 잠금, ENTRY_BLOCKED 뒤 `ObserveStrategyEntryGate` 거절로 다른 레인도 잠김 | design F7 · D3: `dispatch` 가 「계좌 진입 차단」 타입 결과로 주기를 끝냄(오류 아님). `runMarket` FLM 먼저 작성(`function-logic/…runmarket`), BTM 에 두 조립 모두 잠금 0 시험 |
| P1-4 | `max` 의미: 무효 호출자 값은 대체 없이 거절 유지, 측정 실패는 호출자 값으로 떨어지지 않음 | design D2 「바닥 규칙」 · spec 문장 + 시나리오 2 |
| P1-5 (FX 조항) | base 통화 손실은 무환산 · 무권한, 「같은 파도」 = 같은 paired FX collection(같은 observedAt) + precheck 검증 | 델타 문구 갱신 · 승인됨(§2 · §3) |
| P2-1 | 분류 불변 표 시험(`isCentralStrategyIntegrity` · `*strategyScopeRefusal`) + `IssueRefusal.Stage == StageChain` 요구 | design D3 · admit BTM |
| P2-2 | 측정 실패는 봉인된 오류 토큰으로 운반(권한 수집 오류 아님 — kill switch 우선), `MeasureDailyLoss` 는 조립의 마지막 단계 | design D2 |
| P2-3 | Q12 는 포지션 **사이**에서만 참 | spec 문장 축소 · design D1-6 · U8 |
| P2-4 | trade-analytics MODIFIED 델타(성과 행 = Guardian 입력, 부재는 진입만 막음) | `specs/trade-analytics/spec.md` 「분석 경로의 격리」 |
| P2-5 | fail-closed 거부 목록 (a)~(g) + 빈도 추정 + 사람이 잴 읽기 전용 질의 | design 「Fail-closed 가 거부하는 정상 입력」(에이전트는 실 원장에 돌리지 않음) |
| P2-6 | Detail = 시장 · 심볼 · `pos-` 해시 · 사유 / Q7 재강화 함정 | design D2 「Detail」 · D3 |
| P3-1 | US 환산 올림 | design D1-7 · spec 시나리오 |
| P3-2 | KR/US 외 시장은 측정 불가로 통일 | design D1-2 |
| P3-3 | spec SHALL 문구 · 문서 일관성 · archive 는 a136 뒤 | spec · design 머리 |
| P3-4 | tasks 0.4 체크 · 변이 목록 추가 | tasks.md |

P1 다섯의 닫힘은 문서 · RED 수준이다. 코드로 닫히는 것은 GREEN 뒤이고, 닫힘의 독립 재확인은 Manager 몫이다(tasks 0.6.1).

### 4.1 P1 닫힘 독립 재확인 (2026-10-14, 분리된 읽기 전용 컨텍스트 — 판정은 Manager 가 전달)

판정: **BLOCKED** — P1-1 · P1-2 · P1-4 · P1-5 는 문서 · RED 수준에서 CLOSED, **P1-3 NOT-CLOSED**. 수리(대괄호는 Manager 결정 — 각각 좁거나 보수적인 쪽이라 사람 결정 불요):

| 발견 | 내용 | 처분 |
|---|---|---|
| P1-3 핵심 | `runMarket` :935 `if err == nil { continue }` 가 :938 · :966 보다 앞 — nil 주기 반환이 삼킴과 잠금을 모두 피한다(참) | design D3 유지 |
| (b) 새 P1 | 모드 투영 거절은 `ReasonOperatingModeBlocked` 하나(modegate.go:61)이고 기동 복원 실패 latch(mode_projection_wiring.go:33)도 같은 코드, `RejectedError` 는 Reason/Detail 뿐 | [좁게] 조용한 「계좌 진입 차단」은 차단 모드 행의 trigger 가 `ModeTriggerDailyLossLimit` 일 때만(trigger 를 타입으로 운반). 다른 모드 차단 · 기동 복원 실패는 바이트 단위로 오늘과 같음(durable 잠금 포함). 양방향 시험 — design D3-2 · spec · dispatch BTM |
| (a) | 앞서 건너뛴 범위(`*strategyScopeRefusal`, admission.go:99-101)가 있으면 `deliverEachStrategyHandoff` 는 `errors.Join(skipped...)` | [보고 유지] 계좌 차단 결과는 앞선 건너뜀 보고를 바꾸지 않음. spec 시나리오를 「앞서 건너뛴 범위 거절이 없을 때」로 좁히고 혼합 경우를 명시 · 시험(J4 유지) |
| (c) | 경합: 강화 전 승인된 레인 B → `withStrategyEntryGateAuthority` 재확인(strategy_entry_gate_authority.go:62-71) → `ReasonStrategyDispatchFenced`(gateway.go:712-716) → :966 잠금 | [좁게] 펜스 원인이 trigger 일일 손실 모드 변경이면 같은 결과, 아니면 무변경. 열거 · 시험 — design D3-3 · spec · dispatch BTM |
| P2 보호 주문 | 브로커 대기 보호 주문(`protection_sagas` · `protection_mutation_attempts`, protection_orders.go:37-49)을 「수명 안 엔진 매도」에 포함(fail-closed), PLACE 대 CANCEL/AMEND 명시 | design D1-4: SELL PLACE 발송 + 보호 CREATE · REPLACE(PLANNED 아님). CANCEL · 정정 자체는 세지 않음. spec 시나리오 · 거부 목록 (d) |
| P2 (e) | 빈도 문구 오류 — 05:00~09:00 배포 · 재시작 창에서 KST 08:00 뒤에 반영된 US 손실은 그 KR 세션 내내 KR 진입을 막음(KR 동안 US 권한 미봉인) | design D5 · 거부 목록 (e) 정정, 동작 유지(사용자 fail-closed) |
| P2 단일 연결 | journal.go:174 — 색인 범위 읽기(`closed_at` 범위 + 작은 여유), Go 에서 해석, 행을 연 채 네트워크 없음. P3: 문자열 범위 사전 필터가 형식 결함 행을 떨어뜨리면 안 됨 | design D2 「원장 읽기」: 같은 질의에서 NULL · 빈 값 · 비 RFC3339 `closed_at` 의 CLOSED 관리 행 존재를 세어 측정 불가. 지연 측정 task 2.4 · RED 행 추가 |
| §5 | FX 조항은 a136 델타로 옮김(계획 불변) | §5 |

### 4.2 §4.1 수리의 재확인 (2026-10-14, 분리된 읽기 전용 컨텍스트 — 판정은 Manager 가 전달)

판정: **FREEZE-OK** — P1-3 은 문서 · BTM 수준에서 CLOSED, 새 P0/P1 없음.

조건: dispatch BTM 에 이름 붙은 P1-3 RED 시험 — `TestA137ADailyLossModeBlockEndsTheCycleQuietly` · `TestA137OtherModeBlocksBehaveAsBefore`(기동 복원 실패 포함) ·
`TestA137AMixedCycleStillReportsItsSkippedScopes` · BTM 220 의 펜스 시험 둘(`TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock` ·
`TestA137OtherFencesBehaveAsBefore`) — 을 **GREEN 전에 쓰고 실패를 관찰**해야 한다(tasks 1.1).

P3 셋 — Manager 가 모두 수용, **GREEN 에서 적용**:

| P3 | 내용 | 처분 |
|---|---|---|
| 1 | 조용한 경로는 `trigger == ModeTriggerDailyLossLimit` **그리고** `Actor == AUTO` 일 때만(operating_mode.go:245-258 — `Cause` 는 문자열: AUTO 면 열거 trigger, 사람이면 사유 + 승인 참조) | GREEN 에서 적용 — design D3-2 에 조건 추가 |
| 2 | 형식 결함 `closed_at` = 「Z 로 끝나는 정규 RFC3339 UTC 가 아님」(문자열 사전 필터와 같은 정의) | GREEN 에서 적용 — design D2 「원장 읽기」 정의 고정 |
| 3 | 보호 주문 필터 = 인스턴스 수명과 **겹치는** saga · attempt(수명 안에 생성된 것만이 아님) | GREEN 에서 적용 — design D1-4 정정 |

## 5. FX 조항의 자리 — Manager 조정 요청

**결정됨(Manager, 2026-10-14): 조항은 a136 델타에 둔다.** a136 이 문구를 옮겼다(Manager 전달) — a137 델타의 MODIFIED 블록을 지우고 a136 요구를 인용함(이 커밋).

a136 이 같은 조항을 US 수동 매수 금액에 쓰고 a136 이 먼저 archive 한다. 같은 requirement 를 두 델타가 MODIFIED 하면 뒤에 archive 되는 델타가 앞의 것을 통째로
덮는다. **권장: 조항을 a136 델타로 옮기고, a137 델타에서는 그 MODIFIED 를 지운다**(a137 은 조항을 인용만). 옮기기 전까지는 a137 델타에 둔다 — 옮김이 정해지면
a137 쪽 MODIFIED 블록을 지우는 커밋을 하겠다.

## Function Logic Map / Pre-Edit

### Pre-Edit Gate (2026-10-14, 구현 팀메이트 Opus)

```text
Pre-Edit Gate:
- change id / task id: a137-daily-loss-is-measured / 1.1–1.2
- 대상 심볼(편집하는 기존 함수 11 — check_analysis 가 base a000cf29 → 워킹트리로 계산한 목록과 같음):
  execgw.RiskGuardian.IssueEntry · execgw.RiskGuardian.PrecheckQFinalEntry · execgw.EntryGate.ProjectOperatingMode ·
  execgw.Gateway.ObserveStrategyEntryGate · execgw.Gateway.withStrategyEntryGateAuthority · execgw.Gateway.checkEntry ·
  execgw.Gateway.submit(펜스 한 줄) · engine.strategyFirstLegAdmissionBridge.admit · engine.deliverEachStrategyHandoff ·
  시험 fixture execgw_test.qFinalKRRequest · 시험 engine.TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes.
  구조체 칸 추가(함수 아님): RejectedError.cause · EntryGate.modeDailyLossDetail · RiskGuardian.dailyLossRows · QFinalEntryIssuance.DailyLoss ·
  strategyFirstLegGuardian 인터페이스 메서드 둘. 주석: riskcalc.RealizedPnL(TradingDay · AsOf 계약 문장 — Q1).
  새 파일: internal/dailyloss/measure.go · internal/journal/a137_daily_loss_rows.go · internal/execgw/a137_daily_loss.go(+ 시험 5 파일)
- CodeGraph definition/callers/callees/impact: analysis/code-context/(codegraph 1.6.0, R1~R9) + design D0 F1~F14
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/{codegraphcontext-context,evidence-reconciliation}.md
- 기존 동작 파악 근거: HEAD 직접 읽기 + 편집 전 AST(각 번들 ast-base.json, base a000cf29 와 sha 일치 확인) + 기존 시험
  (escalation_test · riskguardian_qfinal_test · strategy_gateway_account_base_test · a112_scope_refusal_census_test · strategy_first_leg_admission_test)
- Function Logic Map / Branch Test Map: analysis/function-logic/ 11 번들(편집 뒤 ast.json + 편집 전 ast-base.json). 무편집 근거 둘(dispatch · runMarket)은
  analysis/ast-evidence/flm-unedited/
- upstream 상속 테스트 영향: no — 진입 판정에 원장 측정 바닥을 더함. 원장에 오늘 닫힌 관리 포지션이 없는 기존 시험은 측정 0 이라 결과 불변;
  q_final fixture 는 토큰을 실어야 함(qFinalKRRequest 한 줄), 승인 사건 열 기대값 하나가 바뀜(measure 추가)
- 실패 테스트 선행 작성: yes — analysis/red/(dailyloss · journal · execgw · engine 컴파일 RED, engine 「조용한 종료」 행동 RED,
  펜스 운반을 되돌린 사본에서의 행동 RED)
- 설정·DB·journal 변경과 rollback: 설정 키 0 · 스키마 0 · 원장 쓰기 0(읽기 질의 셋 추가). 강화는 기존 escalateFor 경로.
  rollback = 이 change 의 커밋 되돌림(데이터 이관 없음)
- 안전 불변식 §0 위반 여부 검토: 통과 — 위험 축소(IssueReduction · evaluateReduction) · flatten · 손절 · 청산 경로는 측정기를 부르지 않음
  (TestA137RiskReductionNeverMeasures). 진입을 더 막는 방향만. 게이트 OFF 면 Guardian 이 없어 측정 없음(engine.go:536).
  precheck 의 원장·브로커 무호출 계약 유지(측정은 승인 함수가 precheck 앞에서, 환산은 봉인 권한 산술뿐). 실 주문 시험 0.
```

**순서 기록(정직하게).** 편집 전 AST 는 편집 대상 9 함수에 대해 편집 **전에** 뽑았다(IssueEntry · PrecheckQFinalEntry 는 a136 1.5 커밋 de9617ab 뒤
커밋된 리비전에서). `checkEntry` · `qFinalKRRequest` · `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes` 셋은 편집 계획에 늦게 들어와
편집 **뒤에** HEAD blob 에서 편집 전 AST 를 재구성했다. 이 Pre-Edit 선언 블록은 편집과 RED 뒤, 첫 Go 커밋 **전에** 썼다 — 선언 내용(대상 · 근거 ·
불변식)은 design D2 · D3 와 freeze 리뷰(§4)에서 편집 전에 고정되어 있었다.

- 새 파일(dailyloss · 원장 읽기 · execgw 측정/표지): Function Logic Map: not-applicable — base 에 같은 qualified 함수가 없는 새 leaf.

## RED

| 영수증 | 내용 |
|---|---|
| `red-dailyloss.txt` | 측정기 순수 함수 12 시험 — 태그 `a137red` 아래 컴파일 RED(`undefined: Closed · Collect`) |
| `red-journal-rows.txt` | 원장 읽기 5 시험 — 구현 파일을 치운 상태의 컴파일 RED(`undefined: DailyLossRow`) |
| `red-execgw.txt` · `red-execgw-testseams.txt` | Guardian · 관문 · 펜스 시험 — 컴파일 RED(`dailyLossRows` · `MeasureDailyLoss` 미정의) |
| `red-engine.txt` | 엔진 시험 — 컴파일 RED(`execgw.DailyLossMeasurement` 미정의) |
| `red-engine-behaviour.txt` | freeze 조건 ① `TestA137ADailyLossModeBlockEndsTheCycleQuietly` **행동 FAIL**(전달 루프 편집 전) — 나머지는 오늘 동작 고정이라 PASS |
| `red-fence-behaviour.txt` | freeze 조건 펜스 시험 — 펜스 운반(`gateway.go` · `strategy_entry_gate_authority.go`)을 HEAD 로 되돌린 사본에서 `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock` **행동 FAIL**, `TestA137OtherFencesBehaveAsBefore` PASS(오늘 동작) |

## 6. 구현 적대적 독립 리뷰 (a1cfe0c2 대상, 분리된 읽기 전용 컨텍스트 — 판정은 Manager 가 전달)

판정: **APPROVE-WITH-FIXES**, P0 없음. 수리는 RED 먼저(새 시험은 지목한 변이에서 실패해야 함 — `analysis/mutation/ledger.md` 3회차).

| 발견 | 내용 | 처분 |
|---|---|---|
| P1-1 | 변이 M2c(`PrecheckQFinalEntry` 의 `Account: account` → `request.Account`) SURVIVED — 살아 있는 문의 바닥을 재는 시험 없음 | `a137_precheck_test.go`: 실제 `PrecheckQFinalEntry` 로 한도 도달 · 측정 불가(타입 표지) · 짝 FX 없는 US 손실 거절 / 있으면 통과(fx=nil 변이도 잡음) |
| P1-2 | 변이 M1b(`dailyLossWindow` 의 롤오버 삭제) SURVIVED — fail-open | `TestA137TheAccountDayWindowHoldsThroughTheRealJournalRead`: 시험 자리 없이 실제 원장 읽기(태그 seam `InsertClosedManagedPositionForTest`)로 네 경계(D 07:59:59 밖 · D 08:00 안 · D+1 07:59:59 안 · D+1 08:00 밖) — 여유(margin) · SQL 범위도 고정 |
| P1-3 | 측정 불가 · 토큰 만료 INPUT_UNAVAILABLE 이 AUTHORITY_MISMATCH 결함 → 서명 활성화 레인 durable 잠금(supervisor :962-966) — design 표 · Q3/Q8 의도와 모순 | **Manager 결정 ①**: 측정기 · 토큰 기원만 타입 표지 `*execgw.DailyLossUnmeasurable`(`dailyLossChainRefusal`), 승인 `DAILY_LOSS_UNMEASURABLE`, 전달 루프 조용히 종료 · 강화 없음 · 진입은 매 주기 재거절. 시험: `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault`(표지 없는 입력 불가는 오늘처럼) · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind`(effective · refreshOnly) · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` |
| P2-1 | 변이 M3b SURVIVED — 단일 latch 조건 | 펜스 시험에 순서 반대(reconcile latch 먼저, 일일 손실 투영 뒤) 경우 추가 — 표지 없음 |
| P2-2 | 일일 손실 모드 차단 동안 query-stale 이 가려짐 | design D3 에 의도로 명시 |
| P2-3 | 자본 ≤ 0 이 살아 있는 문에서 ENTRY_BLOCKED 로 강화됨 | 보수 방향 — spec 문장 · 시나리오 + `TestA137ThePrecheckDoorReportsNoCapitalAsTheDailyLossLimit` |
| P3 | `dailyLossFloor` 의 CompareDecimal 오류 → fail closed | 수리(측정 실패로) |
| P3 | 승인이 잰 토큰을 실어 나르는지 | `TestA137AdmissionCarriesTheMeasuredTokenIntoThePrecheck`(실제 Guardian 토큰, reflect.DeepEqual) |
| P3 | ErrModeAnnouncementFailed 의 시끄러움 | design D3 에 문서화 |
| P3 | 오래 묵은 IN_DOUBT SELL(settled_at NULL) | 거부 목록 (d) 에 추가(끝이 없는 구간 → 겹침, 기존 시험 `in-doubt sell in lifetime`) |
| P3 | `countMalformedClosedAt` 전수 스캔 | tasks 2.4 지연 측정에 포함 |
| P3 | 동등 변이 M6 · M10 · M27b | 리뷰어 판정대로 동등(관측 불가)으로 기록 — ledger.md |

### base 창 문제 (Manager 판단 필요)

base `a000cf29` 뒤에 a136 커밋 셋(844a042c · 1539c675 · 63fd0c92)이 이 change 의 첫 수리 커밋 **앞에** 들어왔다. 그래서 착지를 이 change 의
마지막 커밋에 기록해도 창(base..landing)에 a136 의 수정 함수 13 개가 들어온다(`check_analysis` 2026-10-14 실측: required 24 = a137 11 + a136 13,
a137 11 은 모두 통과). 닫는 길은 WORKFLOW 「사람 승인 base 재고정」(귀속 실측 · 승인 기록 · 단독 커밋)뿐이다 — 이 팀메이트는 base 를 옮기지 않는다.
