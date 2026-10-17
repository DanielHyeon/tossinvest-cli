# Function Logic Map: `admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Source SHA-256: `3c82793300b39454ccac2ff41fe97c0b76ed2190534c5a76c0f9f7abd59652e5`
- Signature: `strategyFirstLegAdmissionBridge.admit(params=2, results=1)`
- Source range: `72:1`–`98:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- admit 은 범위 거절 타입을 만들지도 언급하지도 않는다(census).

## Branches and early returns

- Exact AST return nodes: `75:3, 78:3, 84:3, 87:3, 91:3, 95:3, 97:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 74:2 | 결과 검증 거절 |
| B2 | if | 77:2 | bridge · loader · Guardian 부재 |
| B3 | if | 81:2 | 1차 레그 권한 수집 실패 → `AuthorityCollectionFailed`, 수집 오류를 `cause` 로 운반 |
| B4 | if | 86:2 | 권한 불일치 |
| B5 | if | 90:2 | Guardian precheck 실패 |
| B6 | if | 94:2 | 원자 admission 실패(원장 `BUCKET_USAGE_STALE` 은 cause 없이 문구로 — 타입 없음) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validateStrategyFirstLegResult` | 73:23 |
| `strategyFirstLegRefusal` | 78:10 |
| `b.loader.collectStrategyFirstLegAuthority` | 80:20 |
| `strategyFirstLegRefusal` | 82:14 |
| `err.Error` | 82:98 |
| `validateStrategyFirstLegAuthority` | 86:12 |
| `strategyFirstLegRefusal` | 87:10 |
| `err.Error` | 87:86 |
| `b.guardian.PrecheckQFinalCampaignFirstLeg` | 89:19 |
| `strategyFirstLegRefusal` | 91:10 |
| `err.Error` | 91:86 |
| `b.guardian.IssuePrecheckedQFinalCampaignFirstLeg` | 93:18 |
| `strategyFirstLegRefusal` | 95:10 |
| `err.Error` | 95:90 |
| `string` | 97:81 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(admission). 거절 코드 · 문구 불변, 새로 통과하는 입력 0.
