# Function Logic Map: `LoadCycles`

- Source: `internal/soak/record.go` (272-347)
- Qualified: `LoadCycles`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 9cd4b2d86043c177…) — AST branches 13 · returns 12
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** soak 기록을 읽어 주기 목록을 만든다 — 마지막 줄 찢김만 관용. 호출자 `loadSoakSummary`(attest) · console `readSoak`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 파일 경로 | 호출자 | 없으면 빈 결과 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 274:2 | `if err != nil {` | 무편집 |
| B2 | if at 275:3 | `if errors.Is(err, os.ErrNotExist) {` | 무편집 |
| B3 | if at 300:3 | `if pending == nil {` | 무편집 |
| B4 | if at 304:3 | `if err != nil {` | 무편집 |
| B5 | if at 307:3 | `if c, err = openCycleAccount(c, opener); err != nil {` | 편집 영향 |
| B6 | for at 315:2 | `for scanner.Scan() {` | 무편집 |
| B7 | if at 318:3 | `if len(trimSpace(raw)) == 0 {` | 무편집 |
| B8 | if at 321:3 | `if err := flush(); err != nil {` | 무편집 |
| B9 | if at 326:2 | `if err := scanner.Err(); err != nil {` | 무편집 |
| B10 | if at 330:2 | `if pending != nil {` | 무편집 |
| B11 | if at 332:3 | `if err != nil {` | 무편집 |
| B12 | if at 333:4 | `if errors.Is(err, errUnknownFormat) {` | 무편집 |
| B13 | if at 341:3 | `if c, err = openCycleAccount(c, opener); err != nil {` | 편집 영향 |

Exact AST return positions: 276:4, 278:3, 301:4, 305:4, 308:4, 312:3, 322:4, 327:3, 334:5, 338:4, 342:4, 346:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decodeCycle` | 줄 해독 | 오류면 위치와 함께 반환 | AST calls |
| `openCycleAccount`(a131 신설) | 봉인된 계좌 열기 | 열 수 없으면 빈 계좌(fail-closed) | design D4 |

## State mutations and fallbacks

- 반환 슬라이스. a131: 해독한 주기마다 봉인을 엶(두 자리 — flush 와 마지막 줄). 열 수 없으면 그 줄은 오류이며 마지막 줄 찢김 관용을 받지 않음(R-2a).

## Safety conclusion

- Safe edit boundary: 해독 직후 열기 호출 두 자리. 분기 조건·반환 불변.
- High-risk impact: yes(간접) — attestation 계좌 출처. 왕복 시험이 값 불변을 고정.
