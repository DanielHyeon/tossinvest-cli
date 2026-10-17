# Branch Test Map: `renderFields`

- Source: `internal/obs/notifier.go` (1029-1046) — 편집 뒤. 옛 B1 → 새 B1, 옛 B2 → 새 B3(옛/새 ast.json 정렬), 새 B2 = a133 의 계좌 키 건너뜀.

| Branch | Anchor | Scenario | Test | a133 |
|---|---|---|---|---|
| B1 | range at 1031:2 | 키 모으기 | `TestA133CriticalFieldsLeaveNoAccount` · `TestCriticalAlertIsDurableBeforeItIsSent`(기존) | 편집 영향 |
| B2 | if at 1034:3 | 계좌 키 다섯(`account` · `account_ref` · `account_no` · `masked_account` · `account_seq`, 등록 안 된 값)은 본문에 없고 나머지 키는 남음 | `TestA133CriticalFieldsLeaveNoAccount` · `TestA133BestEffortPathLeavesNoAccount` · `TestA133NormalRelayLeavesNoAccount` | 신설 |
| B3 | range at 1042:2 | 키 순서대로 `k: v` — `position_id: pos-[redacted]` · `symbol` · `reason` · `attempt_id` · `quantity` 줄 | `TestA133CriticalFieldsLeaveNoAccount`(`a133MustKeepMeaning`) | 무편집 |
