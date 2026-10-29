# Function Logic Map: `Client.postWithHeaders`

- Source: `internal/official/client.go` (427-449)
- Qualified function: `Client.postWithHeaders`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 6d145916a6797f3a…)
- AST evidence: `ast.json` — AST branches 3, 반환 4, 호출 9
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

본문을 `json.Marshal` 로 **한 번**(B1, 431) 인코딩하고 makeReq 가 그 바이트로 reader 를 만든다 — `send` 의 401 재전송 본문이 바이트 동일한 근거(R5).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| body | any | PlaceOrder 등 | B1 마샬 오류 |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 432:3, 438:4, 445:3, 448:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 431:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 431 이하 | 미실행 — 패키지 스위트 count=0 |
| B2 | `if at 437:3` | if err != nil { | 아래 「State mutations」 | 소스 줄 437 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `range at 442:3` | for k, v := range extraHeaders { | 아래 「State mutations」 | 소스 줄 442 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

`json.Marshal`, `http.NewRequestWithContext`, `c.send`.

AST 호출 이름 전체: `bytes.NewReader`, `c.send`, `fmt.Errorf`, `http.NewRequestWithContext`, `json.Marshal`, `req.Header.Set`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 편집 계획 없음.
