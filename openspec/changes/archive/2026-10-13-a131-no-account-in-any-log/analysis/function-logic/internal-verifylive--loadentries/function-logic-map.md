# Function Logic Map: `LoadEntries`

- Source: `internal/verifylive/record.go` (511-582)
- Qualified: `LoadEntries`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 9a289d21048578d8…) — AST branches 13 · returns 12
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
| B1 | if at 513:2 | `if err != nil {` | 무편집 |
| B2 | if at 514:3 | `if errors.Is(err, os.ErrNotExist) {` | 무편집 |
| B3 | if at 536:3 | `if pending == nil {` | 무편집 |
| B4 | if at 540:3 | `if err != nil {` | 무편집 |
| B5 | if at 543:3 | `if e, err = openEntryAccount(e, opener); err != nil {` | 편집 영향 |
| B6 | for at 551:2 | `for scanner.Scan() {` | 무편집 |
| B7 | if at 554:3 | `if len(trimSpace(raw)) == 0 {` | 무편집 |
| B8 | if at 557:3 | `if err := flush(); err != nil {` | 무편집 |
| B9 | if at 562:2 | `if err := scanner.Err(); err != nil {` | 무편집 |
| B10 | if at 566:2 | `if pending != nil {` | 무편집 |
| B11 | if at 568:3 | `if err != nil {` | 무편집 |
| B12 | if at 569:4 | `if errors.Is(err, errUnknownFormat) {` | 무편집 |
| B13 | if at 576:3 | `if e, err = openEntryAccount(e, opener); err != nil {` | 편집 영향 |

Exact AST return positions: 515:4, 517:3, 537:4, 541:4, 544:4, 548:3, 558:4, 563:3, 570:5, 573:4, 577:4, 581:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decodeEntry` | 줄 해독 | 오류면 위치와 함께 반환 | AST calls |
| `openEntryAccount`(a131 신설) | 봉인된 계좌 열기 | 열 수 없으면 빈 값(fail-closed) | design D4 |

## State mutations and fallbacks

- 반환 슬라이스. a131: 해독한 줄마다 봉인을 엶(두 자리). 열 수 없으면 오류(마지막 줄 관용 없음, R-2a).

## Safety conclusion

- Safe edit boundary: 해독 직후 열기 호출 두 자리.
- High-risk impact: yes(간접) — 감독 증거 결속 입력.
