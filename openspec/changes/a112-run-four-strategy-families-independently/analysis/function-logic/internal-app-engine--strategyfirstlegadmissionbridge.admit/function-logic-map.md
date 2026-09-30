# Function Logic Map: `admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Source SHA-256: `c31f12fd07855ab32d29c815c8b7b21e14c83add0cd3e1c46503bf01c15eda22`
- Signature: `strategyFirstLegAdmissionBridge.admit(params=2, results=1)`
- Source range: `73:1`–`101:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 거절 타입은 수집 오류에서 **꺼내기만** 한다 — 만들지 않는다(census).

## Branches and early returns

- Exact AST return nodes: `76:3, 79:3, 87:3, 90:3, 94:3, 98:3, 100:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 75:2 | 결과 검증 거절 |
| B2 | if | 78:2 | bridge · loader · Guardian 부재 |
| B3 | if | 82:2 | 1차 레그 권한 수집 실패 → `AuthorityCollectionFailed`(문구 = 수집 오류) |
| B4 | if | 84:3 | **(새)** 수집 오류가 `*strategyScopeRefusal` 이면(errors.As — 타입) 결과에 싣는다; 그 밖의 수집 오류는 싣지 않음(J4) |
| B5 | if | 89:2 | 권한 불일치(편집 전 B4) |
| B6 | if | 93:2 | Guardian precheck 실패(편집 전 B5) |
| B7 | if | 97:2 | 원자 admission 실패(편집 전 B6) — 원장 `BUCKET_USAGE_STALE` 은 여기서 **타입 없이** 올라감 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validateStrategyFirstLegResult` | 74:23 |
| `strategyFirstLegRefusal` | 79:10 |
| `b.loader.collectStrategyFirstLegAuthority` | 81:20 |
| `strategyFirstLegRefusal` | 83:14 |
| `err.Error` | 83:98 |
| `(unnamed)` | 84:15 |
| `errors.As` | 84:45 |
| `validateStrategyFirstLegAuthority` | 89:12 |
| `strategyFirstLegRefusal` | 90:10 |
| `err.Error` | 90:86 |
| `b.guardian.PrecheckQFinalCampaignFirstLeg` | 92:19 |
| `strategyFirstLegRefusal` | 94:10 |
| `err.Error` | 94:86 |
| `b.guardian.IssuePrecheckedQFinalCampaignFirstLeg` | 96:18 |
| `strategyFirstLegRefusal` | 98:10 |
| `err.Error` | 98:90 |
| `string` | 100:81 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(admission). 새로 통과하는 입력 0 — 결과에 분류 값 하나를 더할 뿐, 거절 코드 · 문구는 불변.
