# Branch Test Map: `strategyFirstLegAdmissionBridge.admit`

편집 전 좌표(HEAD `f030e27e`). RED/GREEN 칸은 a137 편집 뒤 갱신한다. 새 갈래(B5-DL)는 편집 뒤 AST 에서 번호가 바뀐다 — 옛/새 ast.json 을 정렬해 재번호할 것(기억 규칙 「분기 번호는 위치다」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 결과 검증 거절 | `TestStrategyFirstLegAdmissionRejectsInvalidBeforeAuthority`(기존) | no | no |
| B2 | 배선 없음 | `TestStrategyFirstLegAdmissionDormantPairedKRUS`(기존) | no | no |
| B3 | 권한 수집 실패 · 원인 보존 | `TestStrategyFirstLegAdmissionClassifiesPairedFailures/*-collection`(기존) | no | no |
| B4 | 권한 결속 실패 | `TestStrategyFirstLegAdmissionRejectsAuthorityTamperPairedKRUS`(기존) | no | no |
| B5 | precheck 거절 → AUTHORITY_MISMATCH, **원인 보존(새)** | `…ClassifiesPairedFailures/*-precheck`(기존) + `TestA137APrecheckRefusalKeepsItsCause`(새) | no | no |
| B5-DL(새) | 일일 손실 한도 거절 → 강화 호출 · `DAILY_LOSS_LIMIT` · cause = `*IssueRefusal` · 범위 거절 아님 | `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault`(새) · `TestA137AFailedEscalationIsJoinedNotLost`(새) | no | no |
| B6·B7 | 버킷 고갈 → 범위 거절 | a112 6.2 시험(기존) | no | no |
| B8 | 원자 발급 실패 | `…ClassifiesPairedFailures/*-atomic`(기존) | no | no |
