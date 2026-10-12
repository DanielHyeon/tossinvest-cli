# Function Logic Map: `NewLogger`

- Source: `internal/obs/log.go` (98-131)
- Qualified: `NewLogger`
- AST evidence: `ast.json` (편집 전(base), `source_sha256` 23d747c66cfeff5e…) — AST branches 6 · returns 3
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** obs.Logger 를 만든다 — 쓰기 대상·레벨·JSON/텍스트·시계·기본 속성을 slog 핸들러로 조립. 생산 호출자 2(`runEngineRun` 엔진 로거 → engine.log, `buildFlattenWiring` → 터미널 stderr).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Writer` | nil 가능 | 호출자 | nil 이면 os.Stderr |
| `opts.Clock` | nil 가능 | 호출자 | nil 이면 clock.System |
| `opts.JSON` | bool | 호출자 | — |
| `opts.Base` | 속성 목록(비어도 됨) | 호출자 | 모든 줄 앞에 붙음 — a131 뒤에는 치환 핸들러를 지남 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 100:2 | `if w == nil {` | 무편집 |
| B2 | if at 104:2 | `if clk == nil {` | 무편집 |
| B3 | if at 114:4 | `if len(groups) == 0 && a.Key == slog.TimeKey {` | 무편집 |
| B4 | if at 122:2 | `if opts.JSON {` | 무편집 |
| B5 | else at 124:9 | `} else {` | 무편집 |
| B6 | if at 127:2 | `if len(opts.Base) > 0 {` | 편집 영향 |

Exact AST return positions: 115:5, 117:4, 130:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `slog.NewJSONHandler`/`NewTextHandler` | 인코딩 | 오류 없음 | AST calls |
| `handler.WithAttrs` | 기본 속성 | 오류 없음 | AST calls |
| `redact.Handler`(a131 신설) | 계좌 키 버림·문자열 치환 | 순수 — 오류 없음 | design D2 |

## State mutations and fallbacks

- 반환 값 외 부작용 없음. a131: 내부 핸들러를 `redact.Handler` 로 감싼 뒤 `Base` 를 붙인다(B6 의 대상이 감싼 핸들러).

## Safety conclusion

- Safe edit boundary: B6 직전에 감싸기 한 줄 추가 — 분기 조건·반환 불변. 줄 내용만 바뀐다(계좌 키 없음·치환).
- High-risk impact: no(직접) — 엔진 로그 바이트만. 주문·청산 판정은 로그를 읽지 않는다.
