# Function Logic Map: `Client.send`

- Source: `internal/official/client.go` (`320`–`366`, 이 change 무편집)
- AST evidence: `ast.json` (`source_sha256` 6d145916a6797f3a…) — branches 9 · returns 8
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 4판(리뷰 4차 P2-6 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 공식 API 요청 한 번(토큰 · 상태 분류).
호출자: `Client.get` 외.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:322` `if err != nil {` | — | — | 무편집 갈래 |
| B2 | `:326` `if err != nil {` | — | — | 무편집 갈래 |
| B3 | `:330` `if err != nil {` | — | — | 무편집 갈래 |
| B4 | `:344` `for attempt := 0; attempt < 2 && code == http.StatusUnauthorized; attempt++ {` | — | — | 무편집 갈래 |
| B5 | `:347` `if err != nil {` | — | — | 무편집 갈래 |
| B6 | `:351` `if err != nil {` | — | — | 무편집 갈래 |
| B7 | `:355` `if err != nil {` | — | — | 무편집 갈래 |
| B8 | `:358` `if !adopted {` | — | — | 무편집 갈래 |
| B9 | `:362` `if code < 200 \|\| code >= 300 {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 근거 자리 참조.

## a130 편집 계획 (4판)

**인용 근거만 · 무편집.** 근거: 429 를 재시도하지 않고 오류로 돌려줌 — 도구는 쓰기 전에 멈춤(design D4 예산, `TestA130ARateLimitedCandleReadFailsAfterOneRequest`).

## Safety conclusion

- Safe edit boundary: 무편집.
- High-risk impact: no(무편집) — 인용 근거.
