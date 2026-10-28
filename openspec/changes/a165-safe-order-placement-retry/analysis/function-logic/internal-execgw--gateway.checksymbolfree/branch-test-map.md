# Branch Test Map: `Gateway.checkSymbolFree`

- Source: `internal/execgw/gateway.go` (802-838; GREEN 뒤 AST 재추출 2026-10-29)
- AST branches 10 — B3 탈출과 B8~B10 latch를 A165 GREEN으로 실행함

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 805:2` | unsettledFor 오류 → 전송 없음 | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B2 | `if at 808:2` | 같은 종목 미종결 attempt → SymbolInFlight(모든 변경) | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B3 | `if at 814:2` | 취소·청산 매도(`raisesExposure=false`) → latch 전에 반환 | A165 GREEN | 계획(D11 RED-b): R3 latch 가 선 종목의 엔진 재취소·청산 매도·flatten 취소가 `checkSymbolFree` 를 통과(latch 질의 호출 0) | `TestA165LatchDoesNotBlockACancel` GREEN |
| B4 | `if at 818:2` | UnresolvedAttempts 오류 → 전송 없음 | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B5 | `range at 821:2` | 미해소 attempt 순회 | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B6 | `if at 823:3` | attemptTargets 오류 | 미측정 | 해당 없음 — 편집 안 함 | 해당 없음 — 편집 안 함 |
| B7 | `if at 826:3` | 같은 종목 UNRESOLVED_IN_DOUBT → 진입 거절 | 미측정 | 계획: 기존 시험이 latch 추가 뒤에도 같은 사유(UnresolvedInDoubt)로 거절됨을 유지(가드 위치 고정) | GREEN 재추출 때 |
| B8 | `if at 832:2` | latch 조회 오류 | A165 GREEN | D11 읽기 오류 → 전송 없음 | 오류 주입 seam 없음 — 기존 submit 오류 경계 유지 |
| B9 | `else at 834:9` | latch 거절 else 갈래 | A165 GREEN | D11 RED-a·c·d: R3 뒤 진입 거절과 journalled r2 해제 | `TestA165EngineCancelLatchBlocksOnlyNewEntryUntilJournalledRelease` GREEN |
| B10 | `if at 834:9` | latch 거절 값 반환 | A165 GREEN | D11 RED-a·c·d: R3 뒤 진입 거절과 journalled r2 해제 | `TestA165EngineCancelLatchBlocksOnlyNewEntryUntilJournalledRelease` GREEN |
