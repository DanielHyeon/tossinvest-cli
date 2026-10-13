# Function Logic Map: `journalTimeStrictlyAfter`

- Source: `internal/journal/fills.go` (`639`–`657`, 이 change 무편집)
- AST evidence: `ast.json` (`source_sha256` 92bc44f947eead2a…) — branches 5 · returns 3
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 4판(리뷰 4차 P2-6 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 원장 시각을 앞선 시각들보다 엄격히 뒤로 만듦(같거나 앞이면 밀어 올림).
호출자: `RecordFill` 외.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:641` `if err != nil {` | — | — | 무편집 갈래 |
| B2 | `:644` `for _, text := range predecessors {` | — | — | 무편집 갈래 |
| B3 | `:645` `if strings.TrimSpace(text) == "" {` | — | — | 무편집 갈래 |
| B4 | `:649` `if parseErr != nil {` | — | — | 무편집 갈래 |
| B5 | `:652` `if !current.After(predecessor) {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 근거 자리 참조.

## a130 편집 계획 (4판)

**인용 근거만 · 무편집.** 근거: 체결 `committed_at` 이 소유 확정 시각보다 엄격히 뒤로 밀려 적힐 수 있음 — 그래서 인스턴스 매도 무결은 committed_at 이 아니라 주문 확정 `settled_at` 을 씀(design D1.5).

## Safety conclusion

- Safe edit boundary: 무편집.
- High-risk impact: no(무편집) — 인용 근거.
