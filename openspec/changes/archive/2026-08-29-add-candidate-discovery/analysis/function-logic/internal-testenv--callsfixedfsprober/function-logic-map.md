# Function Logic Map: `callsFixedFSProber`

- Source: `internal/testenv/static_test.go`
- AST evidence: `ast.json` (revision=current, L145–161, 분기 4개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

이 branch range에서 **새로 생긴** 헬퍼다. `FixedFSProber(`가 주석 밖에 있고 `func `로 시작하지 않는 줄, 즉 **호출**을 찾아 1-based 줄 번호와 함께 보고한다.

존재 이유는 정의 파일의 면제 범위를 좁히는 것이다: 선언은 적을 수 있어야 하지만 같은 파일 안의 호출은 다른 곳의 호출과 같은 결함이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `src string` | Go 소스 전체 | `os.ReadFile` | 빈 문자열이면 `(0, false)` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 줄 순회 | — | — | `TestFixedFSProberIsTestOnly` (B8 경유) |
| B2 | `//` 발견 → 주석 절단 | `code` 축소 | — | 동일 — 주석 안의 언급은 호출이 아니다 |
| B3 | `FixedFSProber(` 없음 | `continue` | — | 동일 |
| B4 | 직전 토큰이 `func` | `continue` | — | 동일 — 선언은 호출이 아니다 |
| (fallthrough) | 호출 발견 | — | `(i+1, true)` | 동일 |
| (끝까지) | 호출 없음 | — | `(0, false)` | 동일 — 현재 두 정의 파일 모두 이 경로 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Split` / `Index` / `TrimSpace` / `HasSuffix` | 줄 단위 어휘 판정 | AST가 아닌 문자열 검사 — 경계는 이 함수의 doc이 적는다 | ast.json calls |

## State mutations and fallbacks

- 순수 함수. 상태 없음.

## Safety conclusion

- Safe edit boundary: 주석 절단과 `func` 접두 판정. 이 둘 중 하나를 잃으면 선언이 호출로 오탐되거나 호출이 선언으로 면제된다.
- High-risk impact: yes (원장 경로 — 내구성) — 이 판정이 느슨해지면 정의 파일 안에서 내구성 가드를 끄는 호출이 통과한다. 테스트 헬퍼이므로 실계좌 부작용은 없다.
