# Branch Test Map: `configServiceFor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch 진입(구조 분기) | 아래 B2/B3 케이스가 함께 덮는다 | no (구조) | yes |
| B2 | `--config-dir` 지정 시 그 디렉터리의 config.json | `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem` (TempDir로 seam 생성 성공) | yes | yes |
| B3 | 미지정 시 기본 경로 | `TestTheSeamSavesAuditsAndPreservesTheFile` | yes | yes |
| B4 | 기본 경로 해석 실패 → nil → 두 seam 모두 미배선 | `TestATypedNilSeamNeverReachesTheInterface` + `TestTheConsoleComesUpWithoutTheLimitsSeam` | yes | yes |
