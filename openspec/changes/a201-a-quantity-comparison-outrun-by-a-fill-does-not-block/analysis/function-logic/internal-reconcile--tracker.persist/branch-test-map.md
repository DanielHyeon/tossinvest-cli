# Branch Test Map: `Tracker.persist` (reconcile) — pre-edit(v3)

- Source: `internal/reconcile/mismatch.go` (878-954). 편집 전 지도 — 커버 측정 · RED/GREEN 은 구현(2.2 · 2.4).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 장부 없음 | 기존 | 해당 없음(편집 전) | 미측정(2.4) |
| B2 | pending 순회 | 기존 | 해당 없음(편집 전) | 미측정(2.4) |
| B3 | 빈 종목 블록 | 기존 | 해당 없음(편집 전) | 미측정(2.4) |
| B4 | 삽입 오류 | 기존 a192 `TestA192AnUnprovenInsertIsNeverBound` | 해당 없음(편집 전) | 미측정(2.4) |
| B5 | 다른 원인 소유 | 기존 | 해당 없음(편집 전) | 미측정(2.4) |
| B6 | Cleared 순회 | 기존 · v3 | 해당 없음(편집 전) | 미측정(2.4) |
| B7 | a192 표시 | a192 `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof` | 해당 없음(편집 전) | 미측정(2.4) |
| B8 | a192 해제 오류 | a192 `TestA192ReleaseWriteFailureKeepsTheGate` | 해당 없음(편집 전) | 미측정(2.4) |
| B9 | a192 재판정 해제 | a192 | 해당 없음(편집 전) | 미측정(2.4) |
| B10 | 크레딧 해제 쓰기 오류 | 기존 · v3(새 원인이 유효 집합에 없을 때의 변이 대상) | 해당 없음(편집 전) | 미측정(2.4) |
| B11 | 활성 행 없음 | 기존 | 해당 없음(편집 전) | 미측정(2.4) |
