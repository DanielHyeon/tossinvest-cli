# a136 evidence reconciliation

| 항목 | CodeGraph | HEAD 읽기 | 결론 |
|---|---|---|---|
| 예약 tx 진입점 | `reserveRows` 호출자 5 | issuance.go:134 · reservations.go:266 · risk_bucket_issuance.go:136 · strategy_first_leg_atomic.go:170 · strategy_lineage.go:358 | 일치 — 훅 한 자리로 세 발급 문 모두 덮음 |
| 발급 문 수 | IssueEntry · IssuePrecheckedQFinalEntry · IssuePrecheckedQFinalCampaignFirstLeg | 셋 모두 OPEN_EXPOSURE 한 종류 예약 | 일치 |
| 한도 스냅샷 digest 사용처 | — | strategy_account_first_leg_authority.go:391 · riskguardian.go:343 · riskguardian_qfinal.go:86 · strategydispatch adapters.go:34 | Limits 5필드 유지(design D6)로 영향 0 |
| 콘솔 seam | `newLimitSettingsSeam` → Save | `gateMembersOf` 가 키 목록을 닫음 | 새 키를 목록에 넣지 않으면 콘솔 저장이 그 키를 건드리지 않음 → 콘솔 분리 시 안전 |

불일치 0.
