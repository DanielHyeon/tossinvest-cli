# Function Logic Map: `dispatch`

- Source: `internal/app/engine/strategy_dispatch_cycle.go`
- Source SHA-256: `72021ef372fe658540d68292cf0556181ad8a981ce81d807eb0903e3b55c67a7`
- Signature: `strategyDispatchCycle.dispatch(params=2, results=2)`
- Source range: `75:1`–`222:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- dispatch 는 범위 거절 타입을 만들지 않는다 — 수집 오류를 나를 뿐.

## Branches and early returns

- Exact AST return nodes: `79:3, 82:3, 88:3, 94:3, 98:3, 116:4, 137:3, 141:3, 145:3, 149:3, 156:4, 158:3, 162:3, 175:3, 181:3, 194:3, 200:3, 204:3, 206:2, 217:5, 220:4`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 78:2 | 80ae96a5 와 같은 분기 |
| B2 | if | 81:2 | 80ae96a5 와 같은 분기 |
| B3 | if | 86:2 | 80ae96a5 와 같은 분기 |
| B4 | if | 93:2 | 80ae96a5 와 같은 분기 |
| B5 | if | 97:2 | 80ae96a5 와 같은 분기 |
| B6 | if | 114:2 | 80ae96a5 와 같은 분기 |
| B7 | if | 115:3 | 80ae96a5 와 같은 분기 |
| B8 | if | 136:2 | 80ae96a5 와 같은 분기 |
| B9 | if | 140:2 | 80ae96a5 와 같은 분기 |
| B10 | if | 144:2 | 80ae96a5 와 같은 분기 |
| B11 | if | 148:2 | 80ae96a5 와 같은 분기 |
| B12 | if | 152:2 | 80ae96a5 와 같은 분기 |
| B13 | if | 155:3 | 수집 오류가 있으면 사슬째 `%w`(문구 = Detail), 없으면 문구만 |
| B14 | if | 161:2 | 80ae96a5 와 같은 분기 |
| B15 | if | 169:2 | 80ae96a5 와 같은 분기 |
| B16 | if | 170:3 | 80ae96a5 와 같은 분기 |
| B17 | if | 174:2 | 80ae96a5 와 같은 분기 |
| B18 | if | 180:2 | 80ae96a5 와 같은 분기 |
| B19 | if | 193:2 | 80ae96a5 와 같은 분기 |
| B20 | if | 199:2 | 80ae96a5 와 같은 분기 |
| B21 | if | 203:2 | 80ae96a5 와 같은 분기 |
| B22 | if | 216:4 | 80ae96a5 와 같은 분기 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `delivered.Result` | 76:12 |
| `validateStrategyFirstLegResult` | 77:23 |
| `errors.New` | 79:28 |
| `errors.New` | 82:28 |
| `StrategyMarket` | 84:12 |
| `cycle.schedule.forMarket` | 85:18 |
| `cycle.fx.forMarket` | 85:52 |
| `errors.New` | 88:28 |
| `strategyFirstLegPlaceIntent` | 93:15 |
| `cycle.gateway.ObserveStrategyProtection` | 96:21 |
| `strings.ToLower` | 96:66 |
| `string` | 96:82 |
| `familyActivation` | 114:19 |
| `cycle.proposals.forMarket` | 114:19 |
| `activation.Verified` | 114:73 |
| `protection.Generation` | 115:6 |
| `activation.ProtectionReadyMinGeneration` | 115:32 |
| `fmt.Errorf` | 116:29 |
| `protection.Generation` | 118:5 |
| `activation.ProtectionReadyMinGeneration` | 118:30 |
| `familyActivation` | 135:12 |
| `cycle.proposals.forMarket` | 135:12 |
| `family.Verified` | 136:5 |
| `errors.New` | 137:28 |
| `family.LeaseCeiling` | 139:23 |
| `cycle.clockNow` | 139:43 |
| `cycle.gateway.ObserveStrategyEntryGate` | 143:25 |
| `strings.ToLower` | 143:69 |
| `string` | 143:85 |
| `cycle.dispatchOwner` | 147:16 |
| `cycle.firstLeg.admit` | 151:14 |
| `fmt.Errorf` | 156:29 |
| `fmt.Errorf` | 158:28 |
| `cycle.journal.LookupDecision` | 160:19 |
| `errors.New` | 162:28 |
| `uint64` | 166:24 |
| `uint64` | 168:20 |
| `strategyOwnerKeyOf` | 169:19 |
| `forScope` | 170:24 |
| `cycle.risk.forMarket` | 170:24 |
| `bundle.Generation` | 171:21 |
| `errors.New` | 175:28 |
| `schedule.restore.Activation.Generation` | 177:26 |
| `schedule.restore.Activation.ExpiresAt` | 178:25 |
| `activationExpiresAt.IsZero` | 180:34 |
| `now.IsZero` | 180:66 |
| `now.Before` | 180:83 |
| `errors.New` | 181:28 |
| `journal.StrategyDispatchMarket` | 183:63 |
| `protection.Generation` | 186:25 |
| `strconv.FormatUint` | 186:68 |
| `protection.Generation` | 186:87 |
| `protection.Digest` | 186:135 |
| `reconciliation.Generation` | 187:29 |
| `reconciliation.Digest` | 187:80 |
| `strategyRuntimeBuildDigest` | 188:94 |
| `min` | 189:9 |
| `activationExpiresAt.Sub` | 189:27 |
| `cycle.journal.IssueVerifiedFirstLegStrategyDispatchLease` | 190:16 |
| `cycle.journal.ClaimStrategyDispatchLease` | 196:18 |
| `strategyFirstLegPlaceIntent` | 202:17 |
| `cycle.gateway.PlaceClaimedStrategy` | 206:9 |
| `cycle.revalidateSchedule` | 216:14 |
| `family.LeaseCeiling` | 219:14 |
| `cycle.clockNow` | 219:34 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(주문 경로). 새로 통과하는 입력 0 — 오류 사슬만 보존.
