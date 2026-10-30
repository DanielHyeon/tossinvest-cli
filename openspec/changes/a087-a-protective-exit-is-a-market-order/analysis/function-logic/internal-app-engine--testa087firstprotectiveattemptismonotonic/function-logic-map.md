# Function Logic Map: `TestA087FirstProtectiveAttemptIsMonotonic`

- Source: `internal/app/engine/a087_lower_limit_internal_test.go`; post-edit AST: `ast.json`; pre-edit AST: `analysis/20261030-daily-refresh/attempt-test-pre.ast.json`.
- Fixture: KR 정규장, 유효 하한가 판, non-nil test admission. 생산 배선과 무관한 시험 전용 값.

| Branch | Condition | Outcome |
|---|---|---|
| B1 | 시간대 로드 오류 | 시험 실패 |
| B2 | 첫 선택이 하한가가 아님 | 시험 실패 |
| B3 | 시도 기록 뒤 관측가 fallback이 아님 | 시험 실패 |

`publish`로 판을 구성하고 `exitLimitPrice` → `markProtectiveAttempt` → `exitLimitPrice` 순으로 확인한다. 외부 호출·실계좌 없음. 수정은 fixture admission 주입뿐이며 생산 계약 변경 없음.
