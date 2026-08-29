# Branch Test Map: `buildCandidateReport`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | not-due 소스 | `cmd/tossctl/candidate_test.go` | — (기존 동작) | yes |
| B2 | 실패한 소스 | 동상 | — (기존 동작) | yes |
| B3 | backoff 중인 소스 | 동상 | — (기존 동작) | yes |
| B4 | 소스별 요청·도착이 JSON과 표에 나온다 | `TestTheJSONReportCarriesBothBlocks` · `TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived` | yes | yes |
| B5 | 소스별 census | `TestTheScanReportAttributesTheRefusalsToASource` · `TestTheJSONReportCarriesBothBlocks` | yes | yes |
| B6 | 사유 맵이 문자열 키로 옮겨진다 | `TestTheJSONReportCarriesBothBlocks`(`not_measured[READING_TRUNCATED] == 3`) | yes | yes |
| B7 | 통과 수가 0이 아니면 note가 바뀐다 | `candidate_review_test.go`의 `structurally 0` 케이스 | — (기존 동작) | yes |
| B8 | code별 위험 건수 | `candidate_test.go` | — (기존 동작) | yes |
| B9 | code별 미측정 건수 | 동상 | — (기존 동작) | yes |
| B10 | 사유 census | 동상 | — (기존 동작) | yes |
| B11 | 가속 미계산 census | 동상 | — (기존 동작) | yes |
| B12 | 밴드 블록 | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` | — (기존 동작) | yes |
| B13 | 밴드별 미측정 | 동상 | — (기존 동작) | yes |
| (Alarms) | 경보가 note **옆에** 실린다 | `TestTheScanOutputSaysSoWhenTheTallyContradictsItself`(JSON `veto.alarms`) · `TestTheOrdinaryNoThresholdScanRaisesNoAlarm`(note는 그대로) | yes | yes |
| (FirstRanksHeld) | 보류 수가 실린다 | `TestTheHeldCountIsRenderedAndSaysWhichCommandCanNeverQualifyAPosition` | yes | yes |
