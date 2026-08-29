# Function Logic Map: `nullable`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `base`, L1816–1822, 분기 1개)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**이다. base(`b268593`) 대비 함수 본문이 byte 동일하며, 이 change가 바로 아래에
`positive`를 삽입하면서 diff hunk가 교차해 evidence가 요구되었다. `ast.json`은 base
revision에서 뜬 것이다.

무엇이 옆으로 들어왔는지는 기록할 가치가 있다. `nullable`은 **문자열**의 부재를 SQL NULL로
보내고, 새로 생긴 `positive`는 **개수**의 부재를 같은 곳으로 보낸다. 둘을 하나로 합치지
않은 이유는 두 타입의 "부재"가 다르기 때문이다 — 문자열은 공백만이어도 부재이고, 개수는
0이 부재다. `nullable("0")`은 "0"을 **보존한다**(진짜 0은 값이다). `positive(0)`은 NULL이다.
같은 함수로 접었다면 저장된 진짜 0 하나가 사라졌을 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | 임의 문자열 | `Reported`의 다섯 문자열 필드 | 공백만이면 NULL |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `strings.TrimSpace(s) == ""` | 없음 | `nil` (SQL NULL) / 아니면 trim된 문자열 | `store_test.go`의 부재·공백 값 케이스 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | present와 usable을 같은 술어로 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 순수 함수, **base와 byte 동일**.
- 옆에 들어온 `positive`가 이 함수를 호출하지도 대체하지도 않는다. `RecordObservations`의 인자 목록에서 두 함수가 나란히 쓰일 뿐이다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입(`positive`)만 존재한다.
- High-risk impact: no (값 정규화). 트리밍이 사라지면 padded 값이 veto 층까지 가서 measured로 읽힌 뒤 파싱에 실패한다 — 함수 주석이 적어 둔 그 세 번째 상태다.
