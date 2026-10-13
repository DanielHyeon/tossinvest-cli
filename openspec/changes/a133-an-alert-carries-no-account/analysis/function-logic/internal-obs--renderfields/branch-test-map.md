# Branch Test Map: `renderFields`

- Source: `internal/obs/notifier.go` (1028-1040, base) — 편집 전. 편집 뒤 좌표는 1.3 에서 갱신.

| Branch | Anchor | Scenario | Test | a133 |
|---|---|---|---|---|
| B1 | range at 1030:2 | 계좌 키 필드는 본문에 안 씀 · 나머지 키는 남음 | `TestA133CriticalFieldsLeaveNoAccount`(신설) | 편집 |
| B2 | range at 1036:2 | 키 순서대로 `k: v` | 기존 패키지 시험 · `TestA133CriticalFieldsLeaveNoAccount` | 무편집 |
