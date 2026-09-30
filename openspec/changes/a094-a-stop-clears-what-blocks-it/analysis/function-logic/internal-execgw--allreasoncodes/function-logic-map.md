# Function Logic Map: `AllReasonCodes`

- Source: `internal/execgw/failclosed.go` (`254`–`311`)
- Qualified: `AllReasonCodes`
- AST evidence: `ast.json` (`source_sha256` fbe97b79db74c5e7…) — 분기 0 · return 1(정렬된 목록)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (없음) | — | 패키지 상수 | 없음 — 목록 리터럴 + 정렬 |

불변식: 이 빌드가 원장 · 알림에 쓰는 reason code 전수이며 골든(`testdata/reason_codes.golden`)과 같아야 한다.

## Branches and early returns

분기 없음. 유일한 경로(B1 행): 리터럴 목록 → `sort.Slice` → 반환. a094 는 목록에 `ReasonOppositePendingOrder` 한 줄을 더했다.

## Calls and live bindings

`sort.Slice`(정렬 비교 함수 리터럴). 원장 · 브로커 호출 0.

## State mutations and fallbacks

없다.

## Safety conclusion

- **Safe edit boundary**: 목록 한 줄 추가. 골든은 생성기(`TOSSOS_UPDATE_GOLDEN=1 go test -run TestWriteReasonCodeGolden`)로만 재생성했다(diff = `opposite_pending_order_exists` 한 줄). a098 의 길이 단언은 `reasonCodesRegisteredAfterA098` 에 이름으로 더했다.
- **High-risk impact**: no — 어휘 열거. 분류 동작은 `classifyMutation` 이 소유.
