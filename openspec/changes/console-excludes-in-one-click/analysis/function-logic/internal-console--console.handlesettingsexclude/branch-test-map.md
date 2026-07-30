# Branch Test Map: `Console.handleSettingsExclude`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | seam 미주입 | `TestExcludeRefusesWhatItCannotDo` | yes | yes |
| B2 | 심볼 없는 요청 | `TestExcludeRefusesWhatItCannotDo` | yes | yes |
| B3 | 판독 불가 config | 저장 0건 단언 | yes | yes |
| B4 | 제외 해제 — 다른 심볼 보존 | `TestReleasingAnExclusion` | yes | yes |
| B5 | 해제 중 저장 거부 | `TestExcludeRefusesWhatItCannotDo` | yes | yes |
| B6 | 신규 제외(멱등·타 필드 보존) | `TestExcludingASymbolFromThePositionsScreen` | yes | yes |
| B7 | 편입 지정된 심볼의 제외 | `TestExcludingADesignatedSymbolDropsTheDesignation` | yes | yes |
| B8 | 엔진이 zeroing할 블록 | `TestExcludeRefusesWhatItCannotDo` | yes | yes |
| — | 손절폭 미침범 | `TestExclusionNeverInventsAStopFraction` | yes | yes |
| — | 대소문자·공백 정규화 | `TestExcludingASymbolIsCaseAndSpaceInsensitive` | yes | yes |
| — | 공지가 현재형 보장을 하지 않는다 | `TestTheExclusionAnswerDefersToTheEngineRestart` | yes | yes |
