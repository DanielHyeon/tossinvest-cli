# Branch Test Map: `refuseStaleBucketUsage`

주석 편집 — 분기 무변. 떠남이 이 함수에 닿는 효과는 a126 시험이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:31` `if len(caps) != len(buckets) {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B2 | `:34` `for i, bucket := range buckets {` | 예 | 기존 — 무편집 | n/a | n/a |
| B3 | `:36` `if err != nil {` | 예 | `TestA126CorruptReceiptedRowsAreUnreadable` 의 판독 불가가 admission 에서 거절로 전파(`TestA126ACorruptDepartedRowLatchesAnUnrelatedActiveOwnerButKeepsItsFill`) | yes | yes |
| B4 | `:43` `if err := latchedUsageRefusal(bucket.Key.Dimension, bucket.Key.Value, usage); err != nil {` | 예 | `TestA126ADepartedRowsLatchFlagStillBlocksEntry` · `TestA126ADepartureDoesNotReleaseAnotherOwnersLatch` — M7 | yes | yes |
| B5 | `:47` `if !ok {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B6 | `:51` `if !ok {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B7 | `:54` `if claimed.Cmp(ledger) < 0 {` | 예 | `TestA126WithoutTheDepartureTheSameEntryIsStale`(떠남 없으면 stale — 양성 대조의 반대편) | yes | yes |
| B8 | `:63` `if caps[i].Key != bucket.Key {` | 예 | 기존 — 무편집 | n/a | n/a |
| B9 | `:67` `if err != nil {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B10 | `:70` `if !found {` | 예 | 기존 — 무편집 | n/a | n/a |
| B11 | `:74` `if !ok {` | 아니오 | 기존 — 무편집 | n/a | n/a |
| B12 | `:77` `if after.Cmp(recorded) > 0 {` | 예 | `TestA126ADepartedRowsLimitStillCaps` — M4 | yes | yes |
