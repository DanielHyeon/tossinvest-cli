# Branch Test Map: `callsFixedFSProber`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 여러 줄 소스를 순회 | `TestFixedFSProberIsTestOnly` (정의 파일 2건에 대해 실행) | yes (헬퍼 부재 시 컴파일 FAIL) | yes |
| B2 | 주석 안의 `FixedFSProber(`는 호출이 아니다 | 동일 — journal.go의 doc 언급이 통과 | yes | yes |
| B3 | 언급 없는 줄은 건너뛴다 | 동일 | no (구조) | yes |
| B4 | `func FixedFSProber(` 선언은 호출이 아니다 | 동일 — 두 정의 파일이 통과 | yes (판정 제거 시 두 정의 파일 모두 FAIL) | yes |
