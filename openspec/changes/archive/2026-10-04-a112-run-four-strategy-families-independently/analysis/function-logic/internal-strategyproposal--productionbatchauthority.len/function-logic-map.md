# Function Logic Map: `ProductionBatchAuthority.Len`

- Source: `internal/strategyproposal/production.go` (150-150)
- Function: `ProductionBatchAuthority.Len` in package `strategyproposal`
- Signature: `ProductionBatchAuthority.Len(params=0, results=1)`
- File SHA-256: `b6e54b502e5092745426f8f4a37e4a02777d525a2099aa90de9f7379ee4a2c18`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

담긴 제안의 수를 돌려준다. 분기도 상태 변경도 없다.

이 번들이 이 lot 에 필요한 이유는 함수 본문이 바뀌어서가 아니라 **이 함수가 사는
구조체가 바뀌어서** 다. 태스크 5.4.3 이 `ProductionBatchAuthority` 에 `absence`·`faulted`
두 필드를 더했고, 그래서 이 메서드의 리시버 타입이 달라졌다. `Len` 이 세는 것은
여전히 `values` 뿐이며 새 필드를 보지 않는다 — 고장 난 배치도 만들어진 제안 수는
그대로 센다. 세는 수에 고장을 섞으면 부르는 쪽이 "제안이 없다"와 "고장이다"를
다시 한 숫자로 뭉치게 되고, 그것이 바로 이 태스크가 없앤 혼동이다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

분기가 없다. 측정 방식과 유일한 행은 `branch-test-map.md` 에 있다.

Exact AST return positions: 150:55.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 150:62 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.

## Safety conclusion

읽기 전용 접근자이며 이 lot 에서 동작이 바뀌지 않았다. 새 필드는 `Fault()` 만 본다.
