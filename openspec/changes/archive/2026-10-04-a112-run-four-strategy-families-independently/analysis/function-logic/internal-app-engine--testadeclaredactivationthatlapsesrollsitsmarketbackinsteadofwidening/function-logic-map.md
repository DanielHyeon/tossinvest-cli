# Function Logic Map: `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening (시험)`

- Source: `internal/app/engine/a112_family_rollback_test.go`
- Source SHA-256: `c1b6888e6bccd1b9ca9e54c43abbae8063f5e9cf12178e98b9170634d4c73979`
- Signature: `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening(params=1, results=0)`
- Source range: `122:1`–`224:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 판정 없음.

## Branches and early returns

- Exact AST return nodes: `135:56, 140:5, 146:5, 152:5, 157:5, 163:5`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | range | 129:2 | 시험 자신의 갈래 |
| B2 | if | 178:4 | 시험 자신의 갈래 |
| B3 | if | 181:4 | 시험 자신의 갈래 |
| B4 | if | 184:4 | 시험 자신의 갈래 |
| B5 | switch | 190:4 | 시험 자신의 갈래 |
| B6 | case | 191:4 | 시험 자신의 갈래 |
| B7 | if | 192:5 | 시험 자신의 갈래 |
| B8 | if | 195:5 | 시험 자신의 갈래 |
| B9 | case | 198:4 | 시험 자신의 갈래 |
| B10 | if | 199:5 | 시험 자신의 갈래 |
| B11 | case | 202:4 | 시험 자신의 갈래 |
| B12 | if | 203:5 | 시험 자신의 갈래 |
| B13 | if | 207:5 | 시험 자신의 갈래 |
| B14 | if | 211:5 | 시험 자신의 갈래 |
| B15 | if | 215:5 | 시험 자신의 갈래 |
| B16 | if | 218:5 | 시험 자신의 갈래 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `time.Date` | 123:9 |
| `writeFamilyActivation` | 139:12 |
| `now.Add` | 139:47 |
| `now.Add` | 139:68 |
| `writeFamilyActivation` | 145:12 |
| `now.Add` | 145:47 |
| `now.Add` | 145:70 |
| `writeFamilyActivation` | 151:12 |
| `now.Add` | 151:47 |
| `now.Add` | 151:68 |
| `strings.Repeat` | 157:87 |
| `writeFamilyActivation` | 162:5 |
| `now.Add` | 162:40 |
| `now.Add` | 162:61 |
| `strings.Repeat` | 163:87 |
| `t.Run` | 167:3 |
| `familyGateFixture` | 168:18 |
| `testStrategyProposalLoader` | 169:13 |
| `entry.setup` | 170:11 |
| `rollbackLoader` | 172:14 |
| `rollbackRoutes` | 174:14 |
| `loader.loadFamilyActivation` | 176:23 |
| `context.Background` | 176:51 |
| `forMarket` | 177:5 |
| `routeReadySchedulePair` | 177:5 |
| `routes.forMarket` | 177:62 |
| `t.Fatalf` | 179:5 |
| `errors.Is` | 181:28 |
| `t.Fatalf` | 182:5 |
| `activation.Verified` | 184:23 |
| `t.Fatalf` | 185:5 |
| `activation.Verified` | 185:49 |
| `forMarket` | 188:17 |
| `a112PairOnly` | 188:17 |
| `loader.collect` | 188:30 |
| `context.Background` | 188:45 |
| `routeReadySchedulePair` | 188:67 |
| `proposalFXPair` | 189:5 |
| `selectedFamily` | 192:15 |
| `t.Fatalf` | 193:6 |
| `t.Fatalf` | 196:6 |
| `selectedFamily` | 199:15 |
| `t.Fatalf` | 200:6 |
| `t.Fatalf` | 204:6 |
| `len` | 205:7 |
| `t.Fatalf` | 208:6 |
| `string` | 210:22 |
| `strings.Join` | 211:8 |
| `strings.Join` | 211:63 |
| `t.Fatalf` | 212:6 |
| `Verified` | 215:8 |
| `authority.familyActivation` | 215:8 |
| `t.Fatal` | 216:6 |
| `Single` | 218:24 |
| `authority.dispatchHandoff` | 218:24 |
| `t.Fatal` | 219:6 |

## State mutations and fallbacks

- 시험 fixture.

## Safety conclusion

- 시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이).
