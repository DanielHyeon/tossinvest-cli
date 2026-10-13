# a130 census — 운영 저널 읽기 전용 사본 (수 · 사유만, F-6)

- 순간: 사본 2026-10-12T21:34:39Z(python `sqlite3` backup, 원본 `mode=ro`), 스키마 32. 측정 2026-10-13, 판정 코드 = a130 3판(커밋 전 트리 밖 사본 —
  GREEN 커밋 뒤 같은 시험으로 다시 재서 sha 를 이 파일에 적음).
- 명령: `A130_JOURNAL_COPY=<사본> go test -count=1 -run TestA130JournalCopyCensus -v ./internal/journal/`

| 항목 | 수 |
|---|---|
| CLOSED | 73 |
| 성과 행 있음 | 31 |
| 동결 행 재계산 = 저장 값(손익 · R · 위험 · 수량) | 31/31 |
| 동결 행 중 a130 판정이 이제 거절하는 것 | 0 |
| 성과 행 없음 | 42 |
| 행 없음 · a130 판정이 계산함 | 3 |
| 행 없음 · `post_adoption_disposal` | 18 |
| 행 없음 · `unmanaged_position` | 12 |
| 행 없음 · `no_intake_observation` | 8 |
| 행 없음 · `sell_unpriced` | 1 |
| **backfill 모집단**(체결로 닫힌 편입 · 행 없음) | 12 |
| 모집단 · fill | 3 |
| 모집단 · refused `no_intake_observation` | 8 |
| 모집단 · refused `sell_unpriced` | 1 |

`post_adoption_disposal` 18 은 전부 모집단 밖(대사로 닫힘 — ADJUSTMENT_CLOSED)이다. 종목 · 수량 · 손익은 저장소에 두지 않는다.
