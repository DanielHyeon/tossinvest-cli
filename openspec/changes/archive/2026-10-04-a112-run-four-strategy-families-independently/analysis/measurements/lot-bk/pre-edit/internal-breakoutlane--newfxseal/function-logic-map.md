# Function Logic Map (편집 전): `NewFXSeal`

- Source: `internal/breakoutlane/types.go`
- Source SHA-256: `4c1cfed18ddeb25329097e3e4710bc58f5300cd7115bc3c9171f1d52542b7512`
- Signature: `NewFXSeal(params=1, results=2)`
- Source range: `153:1`–`164:2`
- AST evidence: `ast.json` — 편집 **전**(a112 B2(breakout 덮개), Manager 판정 2026-10-01 — 수리).

## Inputs and invariants

- 편집 계획: 역방향(instrument→account, 통화 다름) 입력에서 호출자 digest 를 **주어진 모양 그대로** 먼저 검증한 뒤 정규화(비율 뒤집기 · 방향 바꾸기)하고 정규형으로 다시 봉인한다. 편집 전에는 덮어쓴 뒤 자기 비교라 그 방향의 호출자 digest 가 검증되지 않았다(공허 검사 — fail-open 모양).

## Branches and early returns

- Exact AST return nodes: `161:3`, `163:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 155:2 | `if input.Direction == FXInstrumentToAccount && !sameCurrency {` |
| B2 | if | 160:2 | `if input.Direction != FXAccountToInstrument // !canonical(input.AccountCurrency) // !canonical(input.InstrumentCurrency) // input.RateNum == 0 // input.RateDen == 0 // sameCurrency && (input.RateNum != 1 // input.RateDen != 1) // input.Scale == 0 // input.AsOfMS > input.FreshUntilMS // input.Digest != FXSealDigest(input) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `FXSealDigest` | 158:18 |
| `canonical` | 160:50 |
| `canonical` | 160:87 |
| `FXSealDigest` | 160:305 |
| `errors.New` | 161:20 |

## Safety conclusion

- High-risk(FX · 사이징) — 보수 방향: 이전에 수락되던 입력(digest 없거나 틀린 역방향 봉인)을 거절한다. 생산 호출자(패키지 밖) 0 — breakout 미배선이라 생산 효과 0.
