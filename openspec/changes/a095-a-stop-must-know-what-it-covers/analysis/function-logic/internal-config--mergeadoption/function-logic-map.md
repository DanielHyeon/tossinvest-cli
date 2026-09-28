# Function Logic Map: `mergeAdoption`

- Source: `internal/config/engine.go` (`268`–`284`)
- Qualified: `mergeAdoption`
- AST evidence: `ast.json` (`source_sha256` 9d3ab3d2a37da777…)
- Risk scan: `risk-pattern-report.md`
- 분기 3 · return 2 · 호출 3

**역할.** 설정 파일의 편입 블록을 엔진 설정에 옮긴다. 검증에 실패한 블록은 전체를 0으로 만들고 사유를 남긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw` | 파일의 편입 블록 | config.json | B1 — 없으면 기본값 유지 |
| `next.validate()` | 범위 검증 | 정본 exit-policy 「범위 검증」 | B3 — 실패면 `Adoption{Rejected: why}` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/config/ -count=1 -covermode=set`로 만든 `analysis/harness/coverage/r8-config.out` 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:269` `if raw == nil {` | — | :270 | 예 |
| B2 | if | `:273` `if raw.Enabled != nil {` | `normaliseSymbols` | — | 예 |
| B3 | if | `:279` `if why := next.validate(); why != "" {` | `next.validate` | :281 | 예 |

## Calls and live bindings

`normaliseSymbols` · `next.validate`(B3 조건).

결과값이 없다 — 오류를 돌려주지 않는다. 검증 실패는 블록 전체를 0으로 만들고 `Rejected`에 사유를 담는다(B3 창 return).

## State mutations and fallbacks

`cfg.Adoption` — 통과한 블록 또는 `Rejected`만 가진 0 블록.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B3 창이 거부된 블록을 `Adoption{Rejected: why}`로 만든다 — `Enabled` 거짓 · include 없음. 그래서 거부된 편입 블록은 델타의 「운영자가 고른 상태」(편입 꺼짐 ∧ 미지정)와 **모양이 같다** — 8판 델타는 정의에 「설정이 거부되지 않았고」를 넣고 거부는 Q2(a)로 남긴다(r7 F1). **거부된 엔진이 모두 보호를 요청한 엔진은 아니다**(9판 r8 N3): `Adoption.validate`(`config.adoption.validate`) B1 `:161`은 편입 꺼짐 · include 없음 · `DefaultStopPct == 0`일 때만 검증을 건너뛰므로, 의도적으로 끈 블록에 범위 밖 `default_stop_pct`가 남아 있어도 B2 `:164`에서 거부된다. 그 모양은 Q2(a)에 기록한다.
- **High-risk impact**: yes — 편입 여부를 정한다.
