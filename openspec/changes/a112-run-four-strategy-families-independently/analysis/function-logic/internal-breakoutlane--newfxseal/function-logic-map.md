# Function Logic Map: `NewFXSeal`

- Source: `internal/breakoutlane/types.go`
- Source SHA-256: `138a1b686e25353d1f3f6b4f191b626355c457f3e0f3771a004c4afa6f13fe82`
- Signature: `NewFXSeal(params=1, results=2)`
- Source range: `153:1`–`169:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 B2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 역방향 봉인은 주어진 모양 그대로의 digest 가 맞아야 정규화된다 — 그 뒤 정규형으로 다시 봉인한다(fxValid 재검증과 맞물림).
- 생산 호출자(패키지 밖) 0 — breakout 미배선이라 생산 효과 0(grep: `FXSealInput{` · `breakoutlane.NewFXSeal` 패키지 밖 0).

## Branches and early returns

- Exact AST return nodes: `159:4, 166:3, 168:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 155:2 | 역방향(instrument→account) · 통화 다름 → 검증 후 정규화 |
| B2 | if | 158:3 | **(새)** 주어진 모양의 digest 불일치 → 거절(digest 없음 · 다른 봉인 · 정규형 digest · 봉인 뒤 비율/창 변조) |
| B3 | if | 165:2 | 정규형 검증(방향 · 통화 · 비율 · 스케일 · 창 · 재봉인 digest) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `FXSealDigest` | 158:22 |
| `errors.New` | 159:21 |
| `FXSealDigest` | 163:18 |
| `canonical` | 165:50 |
| `canonical` | 165:87 |
| `FXSealDigest` | 165:305 |
| `errors.New` | 166:20 |

## State mutations and fallbacks

- 상태 변경 없음 — 값 검증 · 정규화.

## Safety conclusion

- High-risk(FX · 사이징) — 보수 방향: 이전에 수락되던 입력(digest 없음 · 틀림 · 봉인 뒤 변조된 역방향 봉인)을 거절한다. 수락 집합은 「주어진 모양으로 바르게 봉인된 역방향」 만큼 줄었다.
