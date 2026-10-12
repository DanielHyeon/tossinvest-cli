# Function Logic Map: `LoadEntries`

- Source: `internal/verifylive/record.go` (461-523)
- Qualified: `LoadEntries`
- AST evidence: `ast.json` (편집 전(base), `source_sha256` 9b44a5d8afd7a4dc…) — AST branches 11 · returns 10
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 검증 기록을 읽어 단계 줄 목록을 만든다 — 마지막 줄 찢김만 관용. 호출자 7(attest 감독 증거 · console · verify).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 파일 경로 | 호출자 | 없으면 빈 결과 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 463:2 | `if err != nil {` | 무편집 |
| B2 | if at 464:3 | `if errors.Is(err, os.ErrNotExist) {` | 무편집 |
| B3 | if at 484:3 | `if pending == nil {` | 무편집 |
| B4 | if at 488:3 | `if err != nil {` | 무편집 |
| B5 | for at 496:2 | `for scanner.Scan() {` | 무편집 |
| B6 | if at 499:3 | `if len(trimSpace(raw)) == 0 {` | 무편집 |
| B7 | if at 502:3 | `if err := flush(); err != nil {` | 무편집 |
| B8 | if at 507:2 | `if err := scanner.Err(); err != nil {` | 무편집 |
| B9 | if at 511:2 | `if pending != nil {` | 무편집 |
| B10 | if at 513:3 | `if err != nil {` | 무편집 |
| B11 | if at 514:4 | `if errors.Is(err, errUnknownFormat) {` | 무편집 |

Exact AST return positions: 465:4, 467:3, 485:4, 489:4, 493:3, 503:4, 508:3, 515:5, 518:4, 522:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decodeEntry` | 줄 해독 | 오류면 위치와 함께 반환 | AST calls |
| `openEntryAccount`(a131 신설) | 봉인된 계좌 열기 | 열 수 없으면 빈 값(fail-closed) | design D4 |

## State mutations and fallbacks

- 반환 슬라이스. a131: 해독한 줄마다 봉인을 열어 `AccountRef` 를 채움(두 자리).

## Safety conclusion

- Safe edit boundary: 해독 직후 열기 호출 두 자리.
- High-risk impact: yes(간접) — 감독 증거 결속 입력.
