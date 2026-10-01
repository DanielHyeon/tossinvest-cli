# Function Logic Map (편집 전): `Validate`

- Source: `internal/strategyprojection/model.go`
- Source SHA-256: `0662dc5ab11eda0213bc4e887cdccbb71feb5115bfd5b4627dc71de81090d08f`
- Signature: `Validate(params=1, results=1)`
- Source range: `266:1`–`283:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.3, Manager 판정 Q1~Q3 2026-10-01).

## Inputs and invariants

- 편집 계획: envelope 검사 뒤에 lanes[8] · coordinators[2] 검사(개수 · 고정 순서 · 열쇠 · enum · null 짝)를 더한다.

## Branches and early returns

- Exact AST return nodes: `268:3`, `271:3`, `276:4`, `279:4`, `282:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 267:2 | `if snapshot.SchemaVersion != SchemaVersion // snapshot.GeneratedAt.IsZero() // len(snapshot.Markets) != 2 {` |
| B2 | if | 270:2 | `if err := validateRuntimeIdentity(snapshot.Runtime); err != nil {` |
| B3 | range | 273:2 | `for _, market := range []Market{MarketKR, MarketUS} {` |
| B4 | if | 275:3 | `if !ok // item.Market != market {` |
| B5 | if | 278:3 | `if err := validateMarketProjection(item, snapshot.GeneratedAt); err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `snapshot.GeneratedAt.IsZero` | 267:48 |
| `len` | 267:81 |
| `errors.New` | 268:10 |
| `validateRuntimeIdentity` | 270:12 |
| `fmt.Errorf` | 271:10 |
| `fmt.Errorf` | 276:11 |
| `validateMarketProjection` | 278:13 |
| `fmt.Errorf` | 279:11 |

## Safety conclusion

- High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 쓰기 없음). 편집은 additive 필드만 채우고 기존 시장 레코드 판정은 불변이어야 한다.
