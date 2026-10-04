# Function Logic Map: `dispatch`

- Source: `internal/app/engine/strategy_dispatch_cycle.go`
- Source SHA-256: `d9d29dfcc61759b835bce75e3f500885704f2f27ee9879056cdccbf4f50a5b3b`
- Signature: `strategyDispatchCycle.dispatch(params=2, results=2)`
- Source range: `75:1`–`221:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- dispatch 는 범위 거절을 만들지 않는다 — admit 이 1차 레그 수집 오류에서 꺼낸 값을 나를 뿐(census).

## Branches and early returns

- Exact AST return nodes: `79:3, 82:3, 88:3, 94:3, 98:3, 116:4, 137:3, 141:3, 145:3, 149:3, 155:4, 157:3, 161:3, 174:3, 180:3, 193:3, 199:3, 203:3, 205:2, 216:5, 219:4`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 78:2 | 편집 전과 같은 분기(좌표만) |
| B2 | if | 81:2 | 편집 전과 같은 분기(좌표만) |
| B3 | if | 86:2 | 편집 전과 같은 분기(좌표만) |
| B4 | if | 93:2 | 편집 전과 같은 분기(좌표만) |
| B5 | if | 97:2 | 편집 전과 같은 분기(좌표만) |
| B6 | if | 114:2 | 편집 전과 같은 분기(좌표만) |
| B7 | if | 115:3 | 편집 전과 같은 분기(좌표만) |
| B8 | if | 136:2 | 편집 전과 같은 분기(좌표만) |
| B9 | if | 140:2 | 편집 전과 같은 분기(좌표만) |
| B10 | if | 144:2 | 편집 전과 같은 분기(좌표만) |
| B11 | if | 148:2 | 편집 전과 같은 분기(좌표만) |
| B12 | if | 152:2 | admission 거절 → 주기 오류(편집 전 B12) |
| B13 | if | 154:3 | **(새)** 거절이 범위 거절이면 타입을 `%w` 로 싣는다 — 문구는 같음(J4 ①: 분류는 타입으로) |
| B14 | if | 160:2 | Guardian 결정 세대 없음(편집 전 B13) |
| B15 | if | 168:2 | **(새)** 계보의 범위 키 정규화 |
| B16 | if | 169:3 | **(새)** 그 범위의 준비된 위험 번들에서 세대를 읽음(범위 번들이 없으면 세대 0 → B17 거절) |
| B17 | if | 173:2 | 서명 위험 정책 세대 없음(편집 전 B14) |
| B18 | if | 179:2 | 편집 전과 같은 분기(좌표만, 편집 전 B15) |
| B19 | if | 192:2 | 편집 전과 같은 분기(좌표만, 편집 전 B16) |
| B20 | if | 198:2 | 편집 전과 같은 분기(좌표만, 편집 전 B17) |
| B21 | if | 202:2 | 편집 전과 같은 분기(좌표만, 편집 전 B18) |
| B22 | if | 215:4 | 편집 전과 같은 분기(좌표만, 편집 전 B19) |

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
| `fmt.Errorf` | 155:29 |
| `fmt.Errorf` | 157:28 |
| `cycle.journal.LookupDecision` | 159:19 |
| `errors.New` | 161:28 |
| `uint64` | 165:24 |
| `uint64` | 167:20 |
| `strategyOwnerKeyOf` | 168:19 |
| `forScope` | 169:24 |
| `cycle.risk.forMarket` | 169:24 |
| `bundle.Generation` | 170:21 |
| `errors.New` | 174:28 |
| `schedule.restore.Activation.Generation` | 176:26 |
| `schedule.restore.Activation.ExpiresAt` | 177:25 |
| `activationExpiresAt.IsZero` | 179:34 |
| `now.IsZero` | 179:66 |
| `now.Before` | 179:83 |
| `errors.New` | 180:28 |
| `journal.StrategyDispatchMarket` | 182:63 |
| `protection.Generation` | 185:25 |
| `strconv.FormatUint` | 185:68 |
| `protection.Generation` | 185:87 |
| `protection.Digest` | 185:135 |
| `reconciliation.Generation` | 186:29 |
| `reconciliation.Digest` | 186:80 |
| `strategyRuntimeBuildDigest` | 187:94 |
| `min` | 188:9 |
| `activationExpiresAt.Sub` | 188:27 |
| `cycle.journal.IssueVerifiedFirstLegStrategyDispatchLease` | 189:16 |
| `cycle.journal.ClaimStrategyDispatchLease` | 195:18 |
| `strategyFirstLegPlaceIntent` | 201:17 |
| `cycle.gateway.PlaceClaimedStrategy` | 205:9 |
| `cycle.revalidateSchedule` | 215:14 |
| `family.LeaseCeiling` | 218:14 |
| `cycle.clockNow` | 218:34 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(주문 경로). 새로 통과하는 입력 0 — 추가는 오류에 타입을 싣는 것과 세대 출처를 범위 번들로 좁힌 것뿐.
