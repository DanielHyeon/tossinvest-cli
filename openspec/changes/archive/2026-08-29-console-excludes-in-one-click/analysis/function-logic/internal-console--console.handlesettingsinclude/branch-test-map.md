# Branch Test Map: `Console.handleSettingsInclude`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | seam 미주입 | `TestAnUnwiredSettingsSeamIsExplained` | no | yes |
| B2 | 심볼 없는 요청 | `TestSettingsPostsWithoutCSRFWriteNothing` | no | yes |
| B3 | 판독 불가 config | 기존 seam 오류 테스트 | no | yes |
| B4 | 지정 해제 | `TestRemovingADesignationOnlyAffectsTheFuture` | no | yes |
| B5 | 해제 중 저장 거부 | `TestAnInvalidSaveWritesNothing` | no | yes |
| B6 | 신규 지정(멱등) | `TestDesignatingASymbolFromThePositionsScreen` | no | yes |
| B7 | 제외된 심볼에 도달한 지정 | `TestDesignatingAnExcludedSymbolSaysTheExclusionWins` | yes | yes |
| B8 | 손절폭 미설정 | `TestDesignationAppliesTheDefaultStopFraction` | no | yes |
| B9 | 엔진이 zeroing할 블록 | `TestAnInvalidSaveWritesNothing` | no | yes |
