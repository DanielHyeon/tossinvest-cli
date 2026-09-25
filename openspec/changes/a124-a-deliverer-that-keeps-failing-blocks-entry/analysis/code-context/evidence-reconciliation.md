# Evidence reconciliation — a124

base `4798d399` · codegraph 1.6.0 · AST 번들 `analysis/function-logic/` (추출 `463cc895`, Go 소스 diff 0)

| id | 사실 | CodeGraph 1.6.0 | AST / HEAD | 결론 |
|---|---|---|---|---|
| R1 | `cycle` 의 생산 호출자 | `invokeStrategyCycle` 1 (+시험 10) | `alertdelivery.go:124` `d.cycle(ctx)` 가 유일(grep `\bd\.cycle\(`) | CodeGraph 오탐 1·누락 1(이름 해소). **HEAD 채택**: 호출자 `Run` 하나 |
| R2 | `PendingAlerts` 호출자 | 비테스트 3 | 비테스트 4 — `alertdelivery.go:150` 추가 | CodeGraph 누락(필드 경유 `d.Journal.`). HEAD 채택. 편집(D4 정렬)의 영향 호출자는 4: 실행자·`alertOps.Pending`(운영자 목록)·`Flush`·`Acknowledge` — 셋은 `limit 0` 이라 정렬만 바뀌고 집합은 불변 |
| R3 | `EntryGate.Block` 호출자 | 18 (BlockSymbol 로 해소됨) | 18 자리(`rg '\.Block\('`, 비테스트) — 그중 `ReasonAlertUndelivered` 5 | 수치 18 은 우연 일치, 대상이 다르다. **HEAD 채택**. a124 는 19 번째 자리(실행자)를 더한다 |
| R4 | `EscalateOperatingMode` 호출자 | 수식 3 / 무수식 6 | 6 (`notifier.go:382` · `runtime.go:462` · `exitloop.go:846` · `retry.go:413` · `riskguardian.go:644` · `retry.go:299` 인터페이스 선언) | 무수식 질의가 HEAD 와 일치. a124 는 7 번째 자리를 더한다 |
| R5 | `deliverOne`/`cycle` 의 callee | 원장·publisher 호출 누락 | AST `deliverOne` 호출 22 · `cycle` 호출 7 (FLM 표) | 함수 내부 호출은 **AST 채택**(권위 경계: 함수 내부 = Go AST) |
| R6 | 판정 입력(D1) | — | `SettleResult` 에 attempts 없음(`alert_claim.go:140-145`) | design D1 의 「반환값의 attempts」 는 **현재 API 로 얻을 수 없다** → review F1 |
| R7 | 영향 시험 | 기본 1(Go 0) / filter 854 | 핵심: `a098_*` 15 · `a099_*`/`a096_*` 11 · `internal/obs` 12 | filter 결과 채택. 보존 시험 `a098_the_backlog_does_not_delay_protection_test.go` 2 함수 rc 0(2026-09-26) |

## 편집 차단 여부

도구 간 불일치 R1–R5 는 전부 HEAD grep·AST 로 해소했다 — 그 자체는 편집 차단 사유가 아니다.
**R6 는 해소되지 않았다**: design 이 가리키는 판정 입력이 코드에 없다. proposal-freeze 리뷰(review.md §0)가 REJECT 이므로
design 개정 전까지 1.2(Pre-Edit) 이후로 가지 않는다.
