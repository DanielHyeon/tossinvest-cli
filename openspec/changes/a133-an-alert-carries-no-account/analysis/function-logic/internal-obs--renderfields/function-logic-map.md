# Function Logic Map: `renderFields`

- Source: `internal/obs/notifier.go` (1028-1040, base)
- Qualified: `renderFields`
- AST evidence: `ast.json` (편집 전, `source_sha256` 408579d504089072…) — AST branches 2 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 알림 문맥 필드(`Event.Fields`)를 키 순서대로 `k: v` 줄로 만들어 알림 본문 끝에 붙일 문자열을 돌려줌. 호출자 1(`notificationFor`, CodeGraph) —
`notificationFor` 의 호출자 3(`publishBestEffort` · `deliver` · `NormalRelay.publish` = design D1 P1~P3). `eventKey` · `encodeFields`(outbox payload)는 이 함수를 쓰지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `fields` | 비어 있지 않은 map(호출자가 `len > 0` 일 때만 부름) | 사건 생산자 | 값은 `%v` — 계좌 키(`account` 등) · `position_id` 가 실릴 수 있음 |

## Branches and early returns

| Branch | AST kind | Source line | a133 |
|---|---|---|---|
| B1 | range at 1030:2 | `for k := range fields {` (키 모으기) | 편집 — `redact.IsAccountKey(k)` 이면 건너뜀(새 if) |
| B2 | range at 1036:2 | `for _, k := range keys {` (쓰기) | 무편집 — 계좌 키는 이미 빠짐 |

Exact AST return positions: 1039:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sortStrings` | 안정 순서 | 순수 | AST calls |
| `fmt.Fprintf` | 줄 쓰기 | strings.Builder — 오류 없음 | AST calls |
| `redact.IsAccountKey`(a133 추가) | 계좌 키 판정(a131 규칙) | 순수 | design D3 |

## State mutations and fallbacks

- 지역 슬라이스 · 빌더만. 모든 키가 계좌 키면 빈 문자열 — 호출자는 `TrimSpace(body + "\n" + "")` 로 본문만 남김.

## Safety conclusion

- Safe edit boundary: B1 안 조건 하나. 반환 문자열에서 계좌 키 줄만 빠짐. 전달 판정 · 중복 제거 입력 아님.
- High-risk impact: no(직접) — 알림 본문 바이트만.
