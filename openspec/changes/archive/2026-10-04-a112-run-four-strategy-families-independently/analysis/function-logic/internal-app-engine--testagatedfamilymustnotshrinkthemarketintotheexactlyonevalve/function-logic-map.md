# Function Logic Map: `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve (시험)`

- Source: `internal/app/engine/a112_family_gate_test.go`
- Source SHA-256: `49511cdebc1ba5506737d375b775aa2a533b358701fb531f33f41051f2400fcf`
- Signature: `TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve(params=1, results=0)`
- Source range: `465:1`–`503:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 repin-1e25b3a3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 잠긴 가족이 소유자 범위 하나를 통째로 지우면 시장을 닫는다(정확히-하나 관문 만족 금지) — 그 닫힘이 판정 활성화를 싣는다.

## Branches and early returns

- Exact AST return nodes: ``.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 468:2 | KR 레인 순회 |
| B2 | if | 469:3 | 지속형이 아니면 건너뜀 |
| B3 | if | 472:3 | 비정상 실패로도 안 잠기면 → Fatal |
| B4 | if | 477:2 | 잠근 레인 수 ≠ 1 → Fatal |
| B5 | if | 482:2 | 시장이 열림 → Fatal(고장이 시스템을 관대하게) |
| B6 | if | 487:2 | 사유 ≠ FAMILY_GATE_CLOSED → Fatal |
| B7 | if | 490:2 | 거절 수 ≠ 경로 수 → Fatal |
| B8 | if | 494:2 | 닫힌 시장이 dispatch 에 건넴 → Fatal |
| B9 | if | 499:2 | **(새)** FAMILY_GATE_CLOSED 닫힘이 판정 활성화를 안 실음 → Fatal |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `familyGateFixture` | 466:25 |
| `runtime.lanesFor` | 468:23 |
| `lane.Key` | 469:6 |
| `lane.Fail` | 472:19 |
| `t.Fatal` | 473:4 |
| `t.Fatalf` | 478:3 |
| `collectUnderGate` | 480:15 |
| `t.Fatalf` | 483:3 |
| `len` | 485:68 |
| `t.Fatalf` | 488:3 |
| `t.Fatalf` | 491:3 |
| `Single` | 494:21 |
| `authority.dispatchHandoff` | 494:21 |
| `t.Fatal` | 495:3 |
| `Verified` | 499:6 |
| `authority.familyActivation` | 499:6 |
| `Generation` | 499:49 |
| `authority.familyActivation` | 499:49 |
| `activation.Generation` | 499:94 |
| `t.Fatalf` | 500:3 |
| `Verified` | 501:4 |
| `authority.familyActivation` | 501:4 |
| `Generation` | 501:45 |
| `authority.familyActivation` | 501:45 |
| `activation.Generation` | 501:88 |

## State mutations and fallbacks

- 시험 코드.

## Safety conclusion

- 시험 코드 — 생산 경로 없음.

a112 7.3.1 SHADOW 로트 — 같은 파일의 다른 함수 편집으로 줄만 밀림(본문 불변)
