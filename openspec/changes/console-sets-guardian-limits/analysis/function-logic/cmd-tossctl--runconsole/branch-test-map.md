# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | ctx 없는 호출 | 기존 커버리지 | no (무변경) | yes |
| B2 | verify record 해석 실패 | 기존 커버리지 | no (무변경) | yes |
| B3 | US verify record 해석 실패 | 기존 커버리지 | no (무변경) | yes |
| B4 | soak record 해석 실패 | 기존 커버리지 | no (무변경) | yes |
| B5 | attestation 경로 해석 실패 | 기존 커버리지 | no (무변경) | yes |
| B6 | journal 경로 해석 실패 → 경고 후 계속 | 기존 커버리지 | no (무변경) | yes |
| B7 | 엔진 마커 경로 해석 성공 | 기존 커버리지 | no (무변경) | yes |
| B8 | 엔진 마커 경로 해석 실패 → 경고 후 계속 | 기존 커버리지 | no (무변경) | yes |
