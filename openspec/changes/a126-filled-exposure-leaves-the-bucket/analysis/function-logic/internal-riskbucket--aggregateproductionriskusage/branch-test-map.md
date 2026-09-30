# Branch Test Map: `aggregateProductionRiskUsage`

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:498` `for _, row := range rows {` | 예 | `TestA126AggregateLeavesOnlyReceiptedRows` · `TestA126AggregateKeepsAnActiveOwnersRows` | yes | yes |
| B2 | `:501` `if !filledOK \|\| !heldOK \|\| rowFilled.Sign() < 0 \|\| rowHeld.Sign() < 0 \|\| rowFilled.BitLen() > 256 \|\| rowHeld.BitLen() > 256 \|\|` | — | 기존 행 계약 시험(무편집 조건) | yes | yes |
| B3 | `:510` `if (row.Receipted != 0 \|\| row.DecisionReceipted != 0) && row.OwnerKeyMatches == 0 {` | 예 | `TestA126AggregateRefusesCorruptReceiptedRows`(owner key differs · decision-side receipt only) · `TestA126CorruptReceiptedRowsAreUnreadable`(reservation owner key diverged) — M6b-2 · M6b-3 | yes | yes |
| B4 | `:513` `if row.Receipted != 0 && (row.OwnerReleasedAt == "" \|\| row.OwnerReleasedAt != row.ReceiptReleasedAt \|\| row.State == "HELD" \|\| rowHe…` | 예 | `TestA126AggregateRefusesCorruptReceiptedRows` · `TestA126CorruptReceiptedRowsAreUnreadable` · `TestA126ACorruptDepartedRowLatchesAnUnrelatedActiveOwnerButKeepsItsFill` — M6 · M6b-1 · M6c | yes | yes |
| B5 | `:522` `if !departed {` | 예 | `TestA126AReleasedOwnersFilledLeavesEveryBucket` · `TestA126AnActiveOwnersFilledStaysInEveryBucket` · `TestA126AReleaseMarkWithoutAReceiptDoesNotLeave` · `TestA126AggregateScopeLatchRevertsTheDeparture` · 되돌림 두 경로 — M1 · M1b · M2 · M3 · M7 | yes | yes |
| B6 | `:526` `if filled.BitLen() > 256 \|\| held.BitLen() > 256 {` | 예 | 기존 overflow(무편집) | yes | yes |
