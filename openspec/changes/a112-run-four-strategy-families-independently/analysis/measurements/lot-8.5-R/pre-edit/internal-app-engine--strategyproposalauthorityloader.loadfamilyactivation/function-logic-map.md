# Function Logic Map (편집 전): `loadFamilyActivation`

- Source: `internal/app/engine/strategy_family_activation.go`
- Source SHA-256: `230cc4c84bc3ff2bec2b98caaed10cdec7fd18e46181e3ce995899bbb0ee3492`
- Signature: `strategyProposalAuthorityLoader.loadFamilyActivation(params=5, results=2)`
- Source range: `154:1`–`190:2`
- AST evidence: `ast.json` — 편집 **전**(a112 8.5 응답 로트(HEAD 178cc196)).

## Inputs and invariants

- 편집 계획: 응답 로트 ②(P2-a): getenv 가 nil 이면 공황 대신 ErrProductionFamilyActivationUnavailable 을 돌려준다 — 분기 하나 추가(함수 머리).

## Branches and early returns

- Exact AST return nodes: `183:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 158:2 | `if market == StrategyMarketUS {` |
| B2 | if | 180:2 | `if market == StrategyMarketUS {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyMarketCalibrationDigest` | 168:20 |
| `strategyrouter.LoadProductionFamilyActivation` | 183:9 |
| `strategyRouterMarket` | 184:40 |
| `strings.TrimSpace` | 185:19 |
| `loader.getenv` | 185:37 |
| `strategyRuntimeBuildDigest` | 187:60 |
| `strings.TrimSpace` | 188:21 |
| `loader.getenv` | 188:39 |

## Safety conclusion

- nil getenv 는 손으로 만든 적재기에서만 도달(생산 생성자는 os.Getenv 로 채움); 공황 → collect recover → INTERNAL_FAILURE 였던 것이 관문 되돌림(rolledBack)으로 바뀌어 FX_NOT_READY 사유가 보존된다. 더 닫힌 쪽(되돌림)이라 노출 증가 0.
