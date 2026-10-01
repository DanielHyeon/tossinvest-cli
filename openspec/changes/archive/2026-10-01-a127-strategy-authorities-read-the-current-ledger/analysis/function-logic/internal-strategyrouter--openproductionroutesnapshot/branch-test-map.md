# Branch Test Map: `openProductionRouteSnapshot`

편집 뒤 — RED(`analysis/impl/red.log`) → GREEN → 변이(`analysis/impl/mutation-1.log`). 편집하지 않은 분기는 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:617` `if err := validateProductionRouteJournalFile(journalPath, ownerUID); err != nil {` | 아니오 | 기존 — 파일 검증 · tx 열기 갈래 | n/a | n/a |
| B2 | `:622` `if err != nil {` | 아니오 | 기존 — 파일 검증 · tx 열기 갈래 | n/a | n/a |
| B3 | `:627` `if err != nil {` | 아니오 | 기존 — 파일 검증 · tx 열기 갈래 | n/a | n/a |
| B4 | `:639` `if err := tx.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil {` | 아니오 | 편집된 갈래(판독 실패를 버전 불일치와 분리 — `journal schema unreadable`) · 시험 없음: not-applicable — `PRAGMA user_version` 판독 실패를 만들 결함 주입 seam 이 없음(a066 6.1 구조 시험 선례와 같은 비례 판단), 같은 거절 신원으로 fail-closed | n/a | n/a |
| B5 | `:642` `if version > journalSchema {` | 예 | `TestA127RouteLoaderRefusesANewerLedgerAndSaysSoAtTheBatchBoundary` — S4 · S9b · S12b | yes | yes |
| B6 | `:645` `if version < journalSchema {` | 예 | `TestA127RouteLoaderRefusesAnOlderLedgerAndSaysSoAtTheBatchBoundary` — S5b; 양성 `TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen` — S2 | yes | yes |
| B7 | `:650` `for _, statement := range []string{productionRouteOwnersSQL, productionRouteCampaignSQL} {` | 예 | prepare — `TestA127RouteLoaderRefusesALedgerMissingACampaignColumnEvenWithoutAnActiveOwner` — S8 | yes | yes |
| B8 | `:652` `if err != nil {` | 예 | 같은 시험(campaign 열 삭제 → prepare 실패 거절) | yes | yes |
