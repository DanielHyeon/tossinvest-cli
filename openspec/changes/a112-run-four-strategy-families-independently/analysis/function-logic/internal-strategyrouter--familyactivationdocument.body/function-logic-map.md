# Function Logic Map: `FamilyActivationDocument.body`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`
- Signature: `FamilyActivationDocument.body(params=0, results=2)`
- Source range: `226:1`–`266:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 도구 경로(매니페스트 작성)의 거절이 이유를 말한다 — 검사 자체는 LoadProductionFamilyActivation 하나가 한다(이 함수 머리말).

## Branches and early returns

- Exact AST return nodes: `229:3, 234:4, 251:48, 252:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 228:2 | 서술자 표가 없는 시장 → `%w: market … has no descriptor table` |
| B2 | range | 232:2 | 켤 가족 순회 |
| B3 | if | 233:3 | 모르는 가족 이름 → `%w: on: unknown family …` |
| B4 | range | 239:2 | 표 순회로 네 서술자 생성 |
| B5 | if | 241:3 | 켠 가족이면 ON |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRouteDescriptors` | 227:10 |
| `len` | 228:5 |
| `fmt.Errorf` | 229:44 |
| `family.Known` | 233:7 |
| `fmt.Errorf` | 234:45 |
| `make` | 238:17 |
| `len` | 238:65 |
| `append` | 244:17 |
| `sort.Slice` | 251:2 |
| `Format` | 260:33 |
| `document.ApprovedAt.UTC` | 260:33 |
| `Format` | 261:33 |
| `document.IssuedAt.UTC` | 261:33 |
| `Format` | 262:33 |
| `document.ExpiresAt.UTC` | 262:33 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.
