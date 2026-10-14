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

순서 지시: a136 이 먼저 구현하고 공유 production 파일(`risk.Policy` · `guardian_wiring.go` · `riskguardian_first_leg.go` · a136 이 MODIFY 하는
risk-management 요구)을 소유한다. a137 은 그 커밋 전에는 새 파일만 만든다. base 는 a137 첫 Go 커밋의 부모.

## 3. Q5 스펙 준수 판정 — 남은 사람 승인

**지금 문장 그대로는 준수하지 못하고, 좁은 MODIFIED 하나로 준수한다**(design D5). KR 요청의 일일 손실에 US 손실을 같은 파도의 US 권한으로 환산하면
「같은 authority … aggregate … 다른 환율 사용 금지」에 걸린다. spec 델타가 「KR과 US는 하나의 account-base Guardian 권위를 공유한다」에 일일 손실 환산 조항
(같은 파도 · 공식 원천 · 봉인 권한만, 없으면 fail-closed, 재조회 · 호출자 환율 · 다른 파도 금지, 요청의 sizing/exposure/reservation 권한은 불변)과
시나리오 둘을 더한다. **이 requirement 수정은 사람 승인이 필요하다.** 승인 전에는 그 조항에 의존하는 코드를 쓰지 않는다.

## 4. proposal-freeze 리뷰

(분리된 컨텍스트에 4보이스 리뷰 — 적대적 Eng · 스펙 · 범위 · 운영 — 를 맡겼다. 판정은 아래에 기록한다.)

## Function Logic Map / Pre-Edit

- 분기 근거 = `analysis/ast-evidence/`(12 함수, 읽기 전용 증거) · `analysis/function-logic/internal-app-engine--strategyfirstlegadmissionbridge.admit/`(편집 대상, 편집 전 좌표).
- `riskguardian.go` · `riskguardian_qfinal.go` · `riskguardian_first_leg.go` 의 편집 대상 함수 FLM 은 a136 커밋 뒤 재추출해 쓴다(파일 sha 가 a136 편집으로 바뀜).
- 새 파일(`internal/dailyloss/*`, 원장 읽기 새 파일, Guardian 측정 새 파일): Function Logic Map: not-applicable — base 에 같은 qualified 함수가 없는 새 leaf.
- Pre-Edit 선언은 production 편집 직전(tasks 0.8).

## RED

- `internal/dailyloss/measure_test.go`(빌드 태그 `a137red`, **미커밋** — base 규칙 때문에 첫 Go 커밋을 a136 뒤로 미룸) — 10 시험, base 에서 컴파일 RED
  (`analysis/red/red-dailyloss.txt`: `undefined: Closed · Collect …`). 기본 빌드는 그 디렉터리를 건너뛴다(`go list ./...` rc 0).
