# Function Logic Map: `Adoption.validate`

- Source: `internal/config/engine.go` (`157`–`168`)
- Qualified: `Adoption.validate`
- AST evidence: `ast.json` (`source_sha256` 9d3ab3d2a37da777…)
- Risk scan: `risk-pattern-report.md`
- 분기 2 · return 3 · 호출 3

**역할.** 편입 블록이 쓸 수 있는지 답한다 — 빈 문자열이면 통과, 아니면 거부 사유.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.Enabled` · `a.IncludeSymbols` · `a.DefaultStopPct` | 편입 블록 | config.json | B1 — 셋 다 비면 검증 없이 통과 |
| `exitpolicy.ValidateStopPct`의 답 | pct 범위 | exit-policy 범위 규칙 | B2 — 범위 밖이면 거부 사유 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/config/ -count=1 -covermode=set`로 만든 `analysis/harness/coverage/r8-config.out` 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:161` `if !a.Enabled && len(a.IncludeSymbols) == 0 && a.DefaultStopPct == 0 {` | `len` | :162 | 예 |
| B2 | if | `:164` `if err := exitpolicy.ValidateStopPct(a.DefaultStopPct); err != nil {` | `err.Error`, `exitpolicy.ValidateStopPct` | :165, :167 | 예 |

## Calls and live bindings

`len` · `exitpolicy.ValidateStopPct`(B2 조건) · `err.Error`.

결과는 거부 사유 문자열이다 — 오류를 돌려주지 않고 사유(빈 문자열 = 통과)로 답한다.

## State mutations and fallbacks

없다 — 순수 판정.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B1 `:161`(`if !a.Enabled && len(a.IncludeSymbols) == 0 && a.DefaultStopPct == 0 {`)만 검증을 건너뛴다. 그러므로 편입을 끄고 include도 없지만 `default_stop_pct`가 0이 아닌 범위 밖 값이면 B2에서 **거부**된다 — 「거부된 엔진 = 보호를 요청한 엔진」은 이 모양에서 거짓이다(9판 r8 N3). 이 모양은 Q2(a)(설정 거부의 등급)의 입력으로 기록한다 — Q2(a)=critical이면 의도적으로 끈 엔진이 critical을 받게 되며, 그것은 안전 불변식 3이 지키는 OFF 동등성과 부딪힌다.
- **High-risk impact**: yes — 편입 여부와 거부를 정한다.
