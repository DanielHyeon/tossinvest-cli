# Function Logic Map: `Client.send`

- Source: `internal/official/client.go` (320-366)
- Qualified function: `Client.send`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` 6d145916a6797f3a…)
- AST evidence: `ast.json` — AST branches 9, 반환 8, 호출 8
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — 분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.

## 역할 (a165)

모든 verb 의 인증 요청 흐름. B4 의 for 루프가 **기존 401 재전송**이다(사람 결정 3 이 「확실히 미접수」로 인정한 자리).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| makeReq | 토큰을 받아 요청을 만드는 클로저 | postWithHeaders 등 | 오류면 그대로 반환 |
| ctx | WithSendTracker 가 붙인 httptrace 포함 | submit | — |

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: 323:3, 327:3, 331:3, 348:4, 352:4, 356:4, 363:3, 365:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 322:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 322 이하 | 실행됨 — 패키지 스위트 count=1 |
| B2 | `if at 326:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 326 이하 | 미실행 — 패키지 스위트 count=0 |
| B3 | `if at 330:2` | if err != nil { | 아래 「State mutations」 | 소스 줄 330 이하 | 실행됨 — 패키지 스위트 count=1 |
| B4 | `for at 344:2` | for attempt := 0; attempt < 2 && code == http.StatusUnauthorized; attempt++ { | 아래 「State mutations」 | 소스 줄 344 이하 | 실행됨 — 패키지 스위트 count=1 |
| B5 | `if at 347:3` | if err != nil { | 아래 「State mutations」 | 소스 줄 347 이하 | 미실행 — 패키지 스위트 count=0 |
| B6 | `if at 351:3` | if err != nil { | 아래 「State mutations」 | 소스 줄 351 이하 | 미실행 — 패키지 스위트 count=0 |
| B7 | `if at 355:3` | if err != nil { | 아래 「State mutations」 | 소스 줄 355 이하 | 미실행 — 패키지 스위트 count=0 |
| B8 | `if at 358:3` | if !adopted { | 아래 「State mutations」 | 소스 줄 358 이하 | 실행됨 — 패키지 스위트 count=1 |
| B9 | `if at 362:2` | if code < 200 \|\| code >= 300 { | 아래 「State mutations」 | 소스 줄 362 이하 | 실행됨 — 패키지 스위트 count=1 |

## Calls and live bindings

B4 조건 `attempt < 2 && code == http.StatusUnauthorized` — 401 일 때만 돈다. 루프 안 `c.tm.refresh` 후 `makeReq(tok)` 로 **다시 만든** 요청을 `doRequest` 한다. B8 `!adopted` 면 break — 그래서 401 재전송은 최대 2회(첫 refresh 가 채택 토큰이면 2회, 아니면 1회). postWithHeaders 는 본문을 `json.Marshal` 로 루프 **밖에서 한 번** 인코딩하고 makeReq 가 그 바이트로 `bytes.NewReader` 를 만들므로 재전송 본문(같은 clientOrderId)은 바이트 동일하다(소스 client.go 427–448).

AST 호출 이름 전체: `c.doRequest`, `c.tm.refresh`, `c.tm.token`, `classifyStatus`, `makeReq`, `unwrapAndDecode`.

## State mutations and fallbacks

- 토큰 캐시 갱신(refresh).

## Safety conclusion

- 편집 계획 없음 — a165 는 시험으로 근거만 고정(사람 결정 3). 401 을 받은 POST 가 실제 미접수라는 실측은 없다.
