# Branch Test Map: `writeCandidateTable`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기존 렌더 블록 | `cmd/tossctl/candidate_test.go`의 표 렌더 단언 | — (기존 동작) | yes |
| B2 | 기존 렌더 블록 | `cmd/tossctl/candidate_test.go`의 표 렌더 단언 | — (기존 동작) | yes |
| B3 | 기존 렌더 블록 | `cmd/tossctl/candidate_test.go`의 표 렌더 단언 | — (기존 동작) | yes |
| B4 | 기존 렌더 블록 | `cmd/tossctl/candidate_test.go`의 표 렌더 단언 | — (기존 동작) | yes |
| B5 | 기존 렌더 블록 | `cmd/tossctl/candidate_test.go`의 표 렌더 단언 | — (기존 동작) | yes |
| B6 | 기존 렌더 블록 | `cmd/tossctl/candidate_test.go`의 표 렌더 단언 | — (기존 동작) | yes |
| B7 | 기존 렌더 블록 | `cmd/tossctl/candidate_test.go`의 표 렌더 단언 | — (기존 동작) | yes |
| B8 | 기존 렌더 블록 | `cmd/tossctl/candidate_test.go`의 표 렌더 단언 | — (기존 동작) | yes |
| B9 | 소스별 요청·도착 줄 | `TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived` | yes | yes |
| B10 | `whole`과 `short` 두 라벨 | 동상(100/3 short, 30/30 whole) | yes | yes |
| B11 | 보류 줄과, 보류가 없을 때의 침묵 | `TestTheHeldCountIsRenderedAndSaysWhichCommandCanNeverQualifyAPosition` | yes | yes |
| B12 | 거부 요약 | `candidate_test.go` | — (기존 동작) | yes |
| B13 | 거부 상세 | 동상 | — (기존 동작) | yes |
| B14 | code별 줄 | 동상 | — (기존 동작) | yes |
| B15 | 사유 census | 동상 | — (기존 동작) | yes |
| B16 | 경보 줄 | `TestTheScanOutputSaysSoWhenTheTallyContradictsItself`(ALARM과 숫자) · `TestTheOrdinaryNoThresholdScanRaisesNoAlarm`(음성 대조) | yes | yes |
| B17 | 블록 헤더는 census가 있을 때만 | `TestTheScanReportAttributesTheRefusalsToASource` | yes | yes |
| B18 | 소스별 `0 of 3` / `2 of 2` | 동상 | yes | yes |
| B19 | 소스별 거부 사유 줄 | 동상(`READING_TRUNCATED 3`) | yes | yes |
| B20 | 가속 미계산 | `candidate_test.go` | — (기존 동작) | yes |
| B21 | 밴드 블록 | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` | — (기존 동작) | yes |
| B22 | 밴드가 없는 code | 동상 | — (기존 동작) | yes |
| B23 | 밴드별 미측정 | 동상 | — (기존 동작) | yes |
| B24 | 보존 switch | `candidate_test.go` | — (기존 동작) | yes |
| B25 | 보존 미측정 | 동상 | — (기존 동작) | yes |
| B26 | 보존 busy | 동상 | — (기존 동작) | yes |
| B27 | 보존 보통 | 동상 | — (기존 동작) | yes |
| B28 | 공간 측정됨 | 동상 | — (기존 동작) | yes |
| B29 | 공간 미측정 | 동상 | — (기존 동작) | yes |
