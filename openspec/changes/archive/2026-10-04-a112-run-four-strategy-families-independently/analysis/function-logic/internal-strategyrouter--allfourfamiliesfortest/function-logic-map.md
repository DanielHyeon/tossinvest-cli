# Function Logic Map: `AllFourFamiliesForTest`

- Source: `internal/strategyrouter/production_family_activation_testseam.go` (63-69, base)
- Source SHA-256: `58436a4ec9e3f6893eeeea65dbd187910d3fe4950a958463ebd04ff094fb6d34`
- Signature: `AllFourFamiliesForTest(params=1, results=1)`
- Pinned revision: `base` — this function's body is **unchanged**. The frozen base `1e25b3a31b7109cf7688768000914ce04360a5a9` is pinned because the diff only touches its neighbourhood: a112 7.3.1 appended `FamilyActivationDesiredOnlyForTest` right after it (tossos_testseams build only), and a pure insertion after a function's last line counts as intersecting that function.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 seam(태그 빌드 전용) — 그 시장의 네 레인 ID 를 서술자 표에서 읽어 켠 목록. 본문 불변.

## Branches and early returns

- Exact AST return nodes: `68:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 65:2 | 표 순회 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRouteDescriptors` | 65:22 |

## State mutations and fallbacks

- 새 map 하나를 채워 돌려준다.

## Safety conclusion

- 시험 seam — 생산 빌드에 없음. 본문 불변.
