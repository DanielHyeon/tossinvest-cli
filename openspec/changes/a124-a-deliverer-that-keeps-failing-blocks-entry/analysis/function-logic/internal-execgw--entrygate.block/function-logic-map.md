# Function Logic Map: `EntryGate.Block`

- Source: `internal/execgw/retry.go` (532-539)
- Revision: current — a124 구현 로트(2026-09-27) (편집 없음, 줄만 이동) 재추출; source_sha256 `a9612c2f35817471eee506ee2cfa3686fb297c3326bf1215eacbd162c2574f4c`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력, `analysis/harness/flm_a124.py` 가 다시 만듦)
- Risk scan: `risk-pattern-report.md`
- Extractor counts: AST branches 1 · returns 0 · calls 2
- Exact AST return positions: (none)
- 편집 전 판(base `4798d399`)의 분석은 이 파일의 git 이력에 있고, 분기 대응은 아래 표다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reason` · `detail` | 사유 · 설명 | 호출자 | 이미 있으면 삽입 없음 |

## Branches and early returns

| Branch | AST anchor | Source text at anchor |
|---|---|---|
| B1 | if at 535:2 | `if _, exists := g.latches[reason]; !exists {` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract |
|---|---|---|
| `g.mu.Lock/Unlock` | 게이트 잠금 | map 연산뿐 |

## State mutations and fallbacks

- 없을 때만 삽입하고 그때만 `revision++` — 처음 설명이 남는다(F9). a124 는 이 함수를 편집하지 않았다(위 필드 추가로 줄만 옮겨짐).

## Safety conclusion

- Safe edit boundary: 대조 근거 — 편집 없음.
- High-risk impact: yes — 진입 게이트.
