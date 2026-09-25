# Function Logic Map: `resolveStrategyRuntimeReader`

- Source: `cmd/tossctl/httpapi.go`
- AST evidence: `ast.json` (편집 전 base `8688f74f`, :278–310, 분기 4)
- Risk scan: `risk-pattern-report.md`

**편집하지 않는다(근거 인용 전용).** 콘솔 해석 `resolveConsoleStrategyRuntime` 이 이 판정과 **동치**여야 한다: B1 디렉터리 해석 실패 → nil · B2/B3 descriptor stat 실패(비부재면 경고) → nil · B4 dial 실패 → sentinel · 종단 → client. 동치는 `TestTheConsoleResolutionMatchesTheDaemons` 가 같은 디스크 상태에서 두 해석의 결과 모양을 비교해 고정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | rootOptions | httpapi | 디렉터리 해석 실패 → nil |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 디렉터리 해석 실패 (:281) | 경고 | `(nil,false)` | 콘솔은 engineDir 가 이미 풀린 뒤에만 부른다 — 대응 없음 |
| B2 | descriptor stat 실패 (:289) | — | `(nil,false)` | `TestTheConsoleResolutionMatchesTheDaemons` |
| B3 | 비부재 stat 오류 (:290) | 경고 | — | `TestTheConsoleResolutionMatchesTheDaemons` |
| B4 | dial 실패 (:300) | 경고 | `(sentinel,false)` | `TestTheConsoleResolutionMatchesTheDaemons` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyprojectionrpc.Dial` | 연결 | 200ms probe | AST |

## State mutations and fallbacks

- 편집 없음.

## Safety conclusion

- Safe edit boundary: 편집 없음.
- High-risk impact: no.
