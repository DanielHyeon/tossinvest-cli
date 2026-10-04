# Function Logic Map: `admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- Source SHA-256: `254b4a6abb0d95febd036b0f437829c391e61000fa71b9b9abac1241ee14444c`
- Signature: `strategyFirstLegAdmissionBridge.admit(params=2, results=1)`
- Source range: `74:1`–`110:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 6.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 거절 타입을 만드는 자리는 1차 레그(위험 범위 국소 원인)와 admit(B6 — 버킷 고갈 코드) 둘 — census.
- 진입 관문 관측 거절(R3)은 이 판정이 넓히지 않는다(dispatch 쪽, 타입 없음 그대로).

## Branches and early returns

- Exact AST return nodes: `77:3, 80:3, 86:3, 89:3, 103:3, 107:3, 109:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 76:2 | 결과 검증 거절 |
| B2 | if | 79:2 | bridge · loader · Guardian 부재 |
| B3 | if | 83:2 | 1차 레그 권한 수집 실패 → cause 운반(5.2.2.2) |
| B4 | if | 88:2 | 권한 불일치 |
| B5 | if | 92:2 | Guardian precheck 거절 → `AuthorityMismatch`(문구 = 거절) |
| B6 | if | 98:3 | **(새)** 거절이 `*execgw.QFinalRefusal` 이고 코드가 `BUCKET_CAP_EXHAUSTED` → 그 범위의 정책 결과(결함 아님) |
| B7 | if | 99:4 | **(새)** 범위 키 정규화 → 범위 거절 타입 `strategyScopeRefusal{cause: QFinalRefusal}`(정규화 실패면 결함 그대로 — 봉인이 보장해 도달 불가) |
| B8 | if | 106:2 | 원자 admission(원장 트랜잭션) 실패 — **범위 거절 아님**(같은 파도 둘째 범위 BUCKET_USAGE_STALE · owner 충돌 = 설계 ④ 보호) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validateStrategyFirstLegResult` | 75:23 |
| `strategyFirstLegRefusal` | 80:10 |
| `b.loader.collectStrategyFirstLegAuthority` | 82:20 |
| `strategyFirstLegRefusal` | 84:14 |
| `err.Error` | 84:98 |
| `validateStrategyFirstLegAuthority` | 88:12 |
| `strategyFirstLegRefusal` | 89:10 |
| `err.Error` | 89:86 |
| `b.guardian.PrecheckQFinalCampaignFirstLeg` | 91:19 |
| `strategyFirstLegRefusal` | 93:14 |
| `err.Error` | 93:90 |
| `(unnamed)` | 98:16 |
| `errors.As` | 98:46 |
| `strategyOwnerKeyOf` | 99:21 |
| `b.guardian.IssuePrecheckedQFinalCampaignFirstLeg` | 105:18 |
| `strategyFirstLegRefusal` | 107:10 |
| `err.Error` | 107:90 |
| `string` | 109:81 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk(admission). 새로 **통과**하는 입력 0 — 거절은 그대로 거절이고, 바뀐 것은 그 거절 뒤 같은 주기의 다른 범위가 평가되는가뿐(스펙 「한 family risk bucket 고갈」 — continuation 계속 평가).
