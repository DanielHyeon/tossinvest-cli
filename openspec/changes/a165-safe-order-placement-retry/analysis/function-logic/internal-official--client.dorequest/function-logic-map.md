# Function Logic Map: `Client.doRequest`

- Source: `internal/official/client.go` (191-207)
- Qualified function: `Client.doRequest`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 6d145916a6797f3a…)
- AST evidence: `ast.json` — AST branches 2, 반환 3, 호출 20
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

요청 1회. B1(194) 전송 오류 → status 0 + `ErrTransport` 감쌈(196). B2(201) 본문 읽기 실패 → status 를 함께 반환하지만 `send` 가 버린다(1.3 A2) — 그래서 a165 는 「응답 첫 바이트를 받았는가」를 httptrace `GotFirstResponseByte` 로 따로 잰다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| req | 요청 | send | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 196:3, 203:3, 206:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 194:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 194 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 201:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 201 이하 | 미실행 — 패키지 스위트 count=0 |

## Calls and live bindings

`c.hc.Do`, `io.ReadAll`, `observeAttempt`, `c.rates.record`.

AST 호출 이름 전체: `append`, `c.hc.Do`, `c.rates.record`, `fmt.Errorf`, `io.ReadAll`, `observeAttempt`, `readRateBudget`, `req.Context`, `resp.Body.Close`, `time.Now`.

## State mutations and fallbacks

- rate budget 기록.

## Safety conclusion

- 편집 계획 없음.
