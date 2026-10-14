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

## 5. FX 조항의 자리 — Manager 조정 요청

**결정됨(Manager, 2026-10-14): 조항은 a136 델타에 둔다.** a136 이 문구를 옮겼다는 전달이 오면 a137 델타의 MODIFIED 블록을 지우고 a136 요구를 인용한다. 그때까지는 둔다.

a136 이 같은 조항을 US 수동 매수 금액에 쓰고 a136 이 먼저 archive 한다. 같은 requirement 를 두 델타가 MODIFIED 하면 뒤에 archive 되는 델타가 앞의 것을 통째로
덮는다. **권장: 조항을 a136 델타로 옮기고, a137 델타에서는 그 MODIFIED 를 지운다**(a137 은 조항을 인용만). 옮기기 전까지는 a137 델타에 둔다 — 옮김이 정해지면
a137 쪽 MODIFIED 블록을 지우는 커밋을 하겠다.

## Function Logic Map / Pre-Edit

- 분기 근거 = `analysis/ast-evidence/`(12 함수, 읽기 전용 증거) · `analysis/function-logic/` 셋(편집 전 좌표): `admit`(편집 대상) · `dispatch`(편집 대상 — 주기 끝내기) · `runMarket`(무편집 근거, P1-3).
- `riskguardian.go` · `riskguardian_qfinal.go` · `riskguardian_first_leg.go` 의 편집 대상 함수 FLM 은 a136 커밋 뒤 재추출해 쓴다(파일 sha 가 a136 편집으로 바뀜).
- 새 파일(`internal/dailyloss/*`, 원장 읽기 새 파일, Guardian 측정 새 파일): Function Logic Map: not-applicable — base 에 같은 qualified 함수가 없는 새 leaf.
- Pre-Edit 선언은 production 편집 직전(tasks 0.8).

## RED

- `internal/dailyloss/measure_test.go`(빌드 태그 `a137red`, **미커밋** — base 규칙 때문에 첫 Go 커밋을 a136 뒤로 미룸) — 12 시험(리뷰 P1-1 · P1-2 반영), base 에서 컴파일 RED
  (`analysis/red/red-dailyloss.txt`: `undefined: Closed · Collect …`). 기본 빌드는 그 디렉터리를 건너뛴다(`go list ./...` rc 0).
