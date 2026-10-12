# Function Logic Map: `LoadCycles`

- Source: `internal/soak/record.go` (230-296)
- Qualified: `LoadCycles`
- AST evidence: `ast.json` (편집 전(base), `source_sha256` 583ceb9012f979bd…) — AST branches 11 · returns 10
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
| B1 | if at 232:2 | `if err != nil {` | 무편집 |
| B2 | if at 233:3 | `if errors.Is(err, os.ErrNotExist) {` | 무편집 |
| B3 | if at 256:3 | `if pending == nil {` | 무편집 |
| B4 | if at 260:3 | `if err != nil {` | 무편집 |
| B5 | for at 268:2 | `for scanner.Scan() {` | 무편집 |
| B6 | if at 271:3 | `if len(trimSpace(raw)) == 0 {` | 무편집 |
| B7 | if at 274:3 | `if err := flush(); err != nil {` | 무편집 |
| B8 | if at 279:2 | `if err := scanner.Err(); err != nil {` | 무편집 |
| B9 | if at 283:2 | `if pending != nil {` | 무편집 |
| B10 | if at 285:3 | `if err != nil {` | 무편집 |
| B11 | if at 286:4 | `if errors.Is(err, errUnknownFormat) {` | 무편집 |

Exact AST return positions: 234:4, 236:3, 257:4, 261:4, 265:3, 275:4, 280:3, 287:5, 291:4, 295:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decodeCycle` | 줄 해독 | 오류면 위치와 함께 반환 | AST calls |
| `openCycleAccount`(a131 신설) | 봉인된 계좌 열기 | 열 수 없으면 빈 계좌(fail-closed) | design D4 |

## State mutations and fallbacks

- 반환 슬라이스. a131: 해독한 주기마다 봉인을 열어 `AccountRef` 를 채움(두 자리 — flush 와 마지막 줄).

## Safety conclusion

- Safe edit boundary: 해독 직후 열기 호출 두 자리. 분기 조건·반환 불변.
- High-risk impact: yes(간접) — attestation 계좌 출처. 왕복 시험이 값 불변을 고정.
