# Function Logic Map: `evaluateFresh`

- Source: `internal/breakoutlane/machine.go`
- Source SHA-256: `ab03efc63446a63ab7efdd387c42ec85e1396c606c43c6f8e3d46d8575e18c8f`
- Signature: `evaluateFresh(params=2, results=1)`
- Source range: `36:1`–`128:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 2.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 1.2 반사실 = 입장 문턱이 1.2 였다면 이 setup 에 돌파 봉이 있었는가(design.md 「1.2 반사실의 의미」).
- 기록 전용: `decisionSeal` 은 provenance 의 RVOL 플래그를 포함하지 않는다(전이 · 계보만) — 단계 · 전이 · 거절 · 수량 · 제안 · 봉인 불변(쌍둥이 비교 시험).
- 입장 경로(B6)와 그 안의 1.2 · 2.0 · 2.5 기록 줄은 바이트 그대로 — B7 은 B6 의 break 뒤라 입장 봉에는 닿지 않는다.
- 패키지 밖 소비자 0(grep: `RVOLAt1200000` · `Provenance()` 의 breakout 소비 hit 0) — breakout 미배선이라 생산 효과 0.

## Branches and early returns

- Exact AST return nodes: `80:3, 91:4, 94:4, 102:4, 105:4, 112:3, 115:3, 119:3, 127:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 40:2 | 범위 봉 순회(저항 · 범위 하단) |
| B2 | if | 41:3 | 범위 봉 high 가 저항을 올림 |
| B3 | if | 44:3 | 범위 봉 low 가 하단을 내림 |
| B4 | range | 52:2 | 범위 뒤 봉 순회 |
| B5 | if | 54:3 | high 만 저항 위(첫 touch) |
| B6 | if | 57:3 | 입장 돌파(close buffer · RVOL >= 1.5 · wick) → 반사실 1.2/2.0/2.5 기록 후 break |
| B7 | if | 70:3 | **(새)** 입장 못 한 봉: close buffer · wick 통과 · 1.2 <= RVOL → `RVOLAt1200000` 기록(기록 전용) |
| B8 | if | 74:2 | 입장 돌파 없음 → 범위 단계 결정 |
| B9 | if | 76:3 | 첫 touch 거절 |
| B10 | if | 84:2 | US 시한 |
| B11 | for | 87:2 | 돌파 뒤 봉 순회 |
| B12 | if | 90:3 | 범위 하단 아래 종가 → INVALIDATED |
| B13 | if | 93:3 | `since > timeout` — **도달 불가**(B16 이 같음에서 먼저 반환, 스위트 커버리지 0 실측) · 동작 동등 방어 |
| B14 | if | 96:3 | retest 뒤 reclaim → RECLAIMED · ARMED |
| B15 | if | 101:3 | 거래량 확장 실패 재돌파 → INVALIDATED |
| B16 | if | 104:3 | 시한 정확 → TIMED_OUT |
| B17 | if | 107:3 | retest 허용폭 안 |
| B18 | if | 111:2 | ARMED 아님 → 진행 단계 결정 |
| B19 | if | 114:2 | 호가 거절 → ARMED + typed refusal |
| B20 | if | 118:2 | 사이징 거절 → ARMED + typed refusal |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `string` | 48:89 |
| `string` | 48:114 |
| `b.valueHighAbove` | 54:6 |
| `BreakoutCloseQualifies` | 54:39 |
| `BreakoutCloseQualifies` | 57:6 |
| `append` | 64:20 |
| `string` | 64:42 |
| `string` | 64:57 |
| `BreakoutCloseQualifies` | 70:6 |
| `newDecision` | 75:8 |
| `decisionSeal` | 78:13 |
| `len` | 87:29 |
| `uint64` | 89:12 |
| `newDecision` | 91:11 |
| `appendTransition` | 91:71 |
| `string` | 91:91 |
| `newDecision` | 94:11 |
| `appendTransition` | 94:68 |
| `string` | 94:88 |
| `append` | 97:20 |
| `string` | 97:42 |
| `string` | 97:66 |
| `newDecision` | 102:11 |
| `appendTransition` | 102:71 |
| `string` | 102:91 |
| `newDecision` | 105:11 |
| `appendTransition` | 105:68 |
| `string` | 105:88 |
| `RetestQualifies` | 107:17 |
| `newDecision` | 112:10 |
| `validateQuote` | 114:16 |
| `newDecision` | 115:10 |
| `size` | 117:12 |
| `newDecision` | 119:10 |
| `append` | 121:18 |
| `string` | 121:40 |
| `newDecision` | 122:7 |
| `hashFields` | 125:17 |
| `v.Config.Digest` | 125:67 |
| `decisionSeal` | 126:11 |

## State mutations and fallbacks

- 상태 변경 없음 — 반환하는 Decision 의 provenance 플래그 하나.

## Safety conclusion

- High-risk(돌파 판정 · 증거) — 판정 경로 무변경, 기록만 늘었다. 거절 · 수락 집합 불변(쌍둥이 비교 · 기존 스위트 · 사이징/호가 신탁 GREEN).
