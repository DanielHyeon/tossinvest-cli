# a130 census — 운영 저널 읽기 전용 사본 (수 · 사유만, F-6)

- 순간: 사본 2026-10-12T21:34:39Z(python `sqlite3` backup, 원본 `mode=ro`), 스키마 32. census 시험은 `OpenReadOnly` 로 엶(이관 없음).
- 명령: `A130_JOURNAL_COPY=<사본> go test -count=1 -run TestA130JournalCopyCensus -v ./internal/journal/` — 판정은 체결 트랜잭션과 같은 `prices = nil`.
- 판정 코드: GREEN 커밋 **dbb0681c**(2026-10-13 재측정, 아래 값).

| 항목 | 수 |
|---|---|
| CLOSED | 73 |
| 성과 행 있음 | 31 |
| 동결 행 재계산 = 저장 값(손익 · R · 위험 · 수량) | 31/31 |
| 동결 행 중 a130 판정이 이제 거절하는 것 | 0 |
| 성과 행 없음 | 42 |
| 행 없음 · 체결 트랜잭션 판정(prices = nil) 사유: `needs_intake_price` 11 · `post_adoption_disposal` 18 · `unmanaged_position` 12 · `sell_unpriced` 1 | 42 |
| backfill 모집단(체결로 닫힌 편입 · 행 없음) | 12 (`needs_intake_price` 11 · `sell_unpriced` 1) |
| 판정 소요(prices = nil, 73 종결) | 중앙 129 µs · 최대 2.694 ms (a130 이전 판정: 중앙 1.2 ms · 최대 1.7~2.4 ms, 같은 사본 · 격리 base worktree) |

모집단의 1분봉 판정은 `intake-candles.md`. 종목 · 수량 · 손익은 저장소에 두지 않는다.
